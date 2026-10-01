package filesystem

import (
	"bytes"
	"encoding/json"
 "fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"env/types"
	"env/utils"
)


type Filesystem struct {

	api bool
	root string

	Standards map[string][]string
	Options map[string]any

	Url string
	Username string
	Password string
	Token string

	Cache map[string]*types.Page
	cacheMutex sync.Mutex

	Stack []*types.Patch
	Patch chan *types.Patch
}

func (f *Filesystem) NewDraft(path string, metadata map[string]any) *types.Page {

	return &types.Page {

		Path: path,
		Name: filepath.Base(path),
		Type: "draft",

		Content: []string{},
		Metadata: map[string]any{},

		Og: nil,
		Stage: "draft",

		Children: nil,
		Sorting: "",
	}
}

// CacheEventDraft ensures an event exists in memory without scheduling a write.
func (f *Filesystem) CacheEventDraft(root, id, title string) *types.Page {
	f.cacheMutex.Lock()
	defer f.cacheMutex.Unlock()
	prefix := root + "/>" + id + "$"
	for path, page := range f.Cache {
		if strings.HasPrefix(path, prefix) {
			return page
		}
	}
	path := prefix + title
	page := f.NewDraft(path, nil)
	if f.Cache == nil {
		f.Cache = map[string]*types.Page{}
	}
	f.Cache[path] = page
	return page
}

func (f *Filesystem) EditPage(page *types.Page, newcontent []string) {
	if page == nil || utils.CompareContent(page.Content, newcontent) {
		return
	}
	if page.Stage != "draft" && page.Stage != "move" {
		page.Stage = "edit"
	}
	page.Content = append([]string{}, newcontent...)
}

func (f *Filesystem) RepathPage(page *types.Page, newpath string) error {



	if page == nil { return nil }

	newpath = filepath.ToSlash(filepath.Clean(strings.TrimSpace(newpath)))
	if newpath == "." || filepath.IsAbs(newpath) || newpath == ".." || strings.HasPrefix(newpath, "../") || strings.ContainsAny(newpath, "\n\r\x00") {
		return fmt.Errorf("invalid page path")
	}
	_, newpath, _ = f.standardizePaths(newpath)
	newpath = filepath.ToSlash(newpath)

	oldpath := page.Path
	if newpath == oldpath { return nil }
	f.cacheMutex.Lock()
	defer f.cacheMutex.Unlock()
	if other := f.Cache[newpath]; other != nil && other != page {
		return fmt.Errorf("destination already exists in cache")
	}
	_, _, target := f.standardizePaths(newpath)
	if _, err := os.Stat(target); err == nil {
		return fmt.Errorf("destination already exists on disk")
	} else if !os.IsNotExist(err) { return err }
	if strings.HasPrefix(newpath, oldpath+"/") {
		return fmt.Errorf("cannot move a page inside itself")
	}

	if parent, ok := f.Cache[filepath.Dir(oldpath)]; ok && parent != page {
		parent.Children = slices.DeleteFunc(parent.Children, func(c *types.Page) bool {
			return c == page
		})
	}

	var walk func(p *types.Page)

	walk = func(p *types.Page) {

		if p == nil { return }

		np := newpath + strings.TrimPrefix(p.Path, oldpath)

		delete(f.Cache, p.Path)
		p.Path = np
		p.Name = filepath.Base(np)

		if p.Stage != "draft" {
			p.Stage = "move"
		}


		f.Cache[np] = p

		for _, child := range p.Children {
			walk(child)
		}
	}

	walk(page)

	if parent, ok := f.Cache[filepath.Dir(newpath)]; ok && parent != page {

		parent.Children = append(parent.Children, page)
		sortChildren(parent.Sorting, parent.Children)

	}
	return nil
}

func (f *Filesystem) DoomPage(page *types.Page) {

	page.Stage = "doom"
}

func (f *Filesystem) NewLoading() *types.Loading {

	return &types.Loading {

		Api: false,
		Mode: "",
		Path: "",

		Meta: true,
		Cont: true,
		Depth: -1,

		Sort: "",
		Filters: []string{} ,
		Latest: -1,

		Token: "",
	}
}

func (f *Filesystem) NewSyncing() *types.Syncing {

	return &types.Syncing {

		Api: false,
		Branch: nil,
		Hard: true,

		Filters: []string{},
	}
}

var errorPage = &types.Page {

	Path: "error",
	Type: "error",
	Name: "error",

	Content: []string{"error"},
	Metadata: map[string]any{"error": "error"},

	Og: nil,
	Stage: "error",

	Children: nil,
	Sorting: "error",
}

func (f *Filesystem) newPatch() *types.Patch {

	return &types.Patch {

		Description: [][]string{},

		Commands: []func(){},
		Timer: nil,
		Done: make(chan struct{}),

	}
}

func (f *Filesystem) newPage(path string, metadata map[string]any) *types.Page {

	return &types.Page {

		Path: path,
		Type: "",
		Name: filepath.Base(path),

		Content: []string{},
		Metadata: map[string]any{},

		Og: nil,
		Stage: "",

		Children: nil,
		Sorting: "",
	}
}

func CreateFilesystem(api bool) *Filesystem {

	f := Filesystem {

		api: api,

		Standards: map[string][]string{},
		Options: map[string]any{},

		Cache: map[string]*types.Page{},

		Stack: []*types.Patch{},
		Patch: make(chan *types.Patch),
	}

	f.reload()

	if !api { 

		ok := f.auth()

		p := f.newPatch()

		p.Commands = []func(){}
		if ok { 
			p.Description = [][]string{[]string{"authentication succeeded!"}}
		} else {
			p.Description = [][]string{[]string{"authentication failed..."}}	
		}

		f.Stack = append(f.Stack, p)
		f.schedulePatch(p)

		req := f.NewLoading()
		req.Api = false
		req.Cont = false
		req.Depth = -1
		req.Filters = []string{}
		req.Latest = -1
		req.Mode = "ghost"
		req.Meta = false
		req.Path = "."
		req.Sort = ""
		req.Token = ""

		f.Load(req) 
	}

	return &f
}

func (f *Filesystem) auth() bool {

	body, err := json.Marshal(map[string]string{
		"username": f.Username,
		"password": f.Password,
	})
	if err != nil { return false }

	httpreq, err := http.NewRequest(http.MethodPost, f.Url+"/auth", bytes.NewReader(body))
	if err != nil { return false }

	httpreq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpreq)
	if err != nil { return false }

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK { return false }

	var out struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil { return false }
	if out.Token == "" { return false }

	f.Token = out.Token ; return true
}

func (f *Filesystem) reload() {

	lines := f.getContent(filepath.Join(f.whereIsRoot(), "proj/.@project/..template"))
	lines, _ =  utils.SplitLines(lines)

	std := map[string][]string{}
	for _, line := range lines {

		key, value := utils.SplitTwo(line, ":")
		_, ok := std[key]
		if ok {
			std[key] = append(std[key], value)
		} else {
			std[key] = []string{value}
		}

		
	}
	
	f.Standards = std

	lines = f.getContent(filepath.Join(f.whereIsRoot(), ".options"))
	lines, _ =  utils.SplitLines(lines)

	for _, line := range lines {

		key, value := utils.SplitTwo(line, ":")

		if key == "url" { f.Url = value }
		if key == "username" { f.Username = value }
		if key == "password" { f.Password = value }
	}
	

	f.Options = f.getMetadata(filepath.Join(f.whereIsRoot(), ".options"))
}


func (f *Filesystem) Load(req *types.Loading) chan *types.Page {

	channel := make(chan *types.Page)

	go func() {

		defer close(channel)

		var ptr *types.Page

		switch req.Api {
		case false:
			_, ptr = f.LoadLocal(req)
		case true:
			_, ptr = f.LoadApi(req)
		}

		if ptr == nil {
			return
		}

		f.merge(ptr)
		ptr.Token = req.Token
		channel <- ptr
	}()

	return channel
}


func (f *Filesystem) LoadLocal(req *types.Loading) (bool, *types.Page) {

	root, relpath, abspath := f.standardizePaths(req.Path)

	collector := &pageCollector{Pages: map[string]*types.Page{}}

	switch f.checkPath(abspath) {

	case "":
		if _, err := os.Stat(abspath); !req.CreateDraft || !os.IsNotExist(err) {
			return false, errorPage
		}
		f.cacheMutex.Lock()
		defer f.cacheMutex.Unlock()
		if page := f.Cache[relpath]; page != nil && page.Stage == "draft" {
			return true, page
		}
		page := f.NewDraft(relpath, nil)
		page.Content = []string{""}
		page.Sorting = req.Sort
		if f.Cache == nil {
			f.Cache = map[string]*types.Page{}
		}
		f.Cache[relpath] = page
		return true, page
		
	case "file":
		info, err := os.Stat(abspath)
		if err != nil { return false, errorPage }

		entry := fs.FileInfoToDirEntry(info)
		err = collector.processEntries(f, abspath, entry, nil, root, abspath, req)
		if err != nil { return false, errorPage }

	case "dir":
		err := filepath.WalkDir(abspath, func(path string, d fs.DirEntry, err error) error {
			return collector.processEntries(f, path, d, err, root, abspath, req)
		})
		if err != nil { return false, errorPage }	
	}

	pages := f.applyFilters(req.Filters, collector.Pages)
	if rootPage := pages[relpath]; rootPage != nil && rootPage.Type == "deep" {
		page := buildPointer(pages)
		if req.Latest >= 0 && len(page.Children) > req.Latest {
			page.Children = page.Children[:req.Latest]
		}
		return true, page
	}
	if req.Latest >= 0 {
		pages = f.filter("latest:"+strconv.Itoa(req.Latest), pages)
	}
	return true, buildPointer(pages)
}

func buildPointer(pages map[string]*types.Page) *types.Page {


	for _, page := range pages {

		parentpath := filepath.Dir(page.Path)
		parent, ok := pages[parentpath]

		if !ok || parentpath == page.Path { continue }

		parent.Children = append(parent.Children, page)
	}

	depth := -1
	var root *types.Page

	for _, page := range pages {

		d := utils.PathDepth(page.Path)

		if root == nil || d < depth {

			depth = d
			root = page
		}
	}

	for _, page := range pages {
		
		sortChildren(page.Sorting, page.Children)
	}

	return root
}

func (f *Filesystem) merge(page *types.Page) {

    if page == nil { return }

    f.cacheMutex.Lock()
    f.Cache[page.Path] = page
    f.cacheMutex.Unlock()

    for _, child := range page.Children {
        f.merge(child)
    }
}

func (f *Filesystem) LoadApi(req *types.Loading) (bool, *types.Page) {

	body, err := json.Marshal(req)
	if err != nil { return false, errorPage }

	httpreq, err := http.NewRequest(http.MethodPost, f.Url+"/load", bytes.NewReader(body))
	if err != nil { return false, errorPage	}

	httpreq.Header.Set("Authorization", "Bearer "+f.Token)
	httpreq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpreq)
	if err != nil {	return false, errorPage }

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK { return false, errorPage }

	page := &types.Page{}
	err = json.NewDecoder(resp.Body).Decode(page)
	if err != nil { return false, errorPage }

	f.restoreOg(page)
	return true, page

}

func (f *Filesystem) restoreOg(page *types.Page) {

	if page == nil { return }

	page.Og = f.snapPage(page)

	for _, child := range page.Children {
		f.restoreOg(child)
	}
}

func (f *Filesystem) snapPage(page *types.Page) *types.Page {

	return &types.Page{

		Path:     page.Path,
		Type:     page.Type,
		Name:     page.Name,

		Content:  page.Content,
		Metadata: page.Metadata,

		Og:       nil,
		Stage:    "",

		Children: nil,
		Sorting:  page.Sorting,
	}
}

type pageCollector struct { Pages map[string]*types.Page }
func (c *pageCollector) processEntries(f *Filesystem, currentpath string, d fs.DirEntry, err error, root string, startabspath string, req *types.Loading) error {

	if err != nil { return err }
	if d == nil { return nil }

	if currentpath != startabspath && (d.Name() == "index" || d.Name() == ".options") {
		if d.IsDir() {
			return filepath.SkipDir
		}
		return nil
	}

	reltostartpath, _ := filepath.Rel(startabspath, currentpath)
	reltorootpath, _ := filepath.Rel(root, currentpath)

	depth := utils.PathDepth(reltostartpath)
	if req.Depth != -1 && depth > req.Depth {
		if d.IsDir() {
			return filepath.SkipDir
		}
		return nil
	}

	page := f.newPage(req.Path, map[string]any{})

	page.Name = d.Name()
	page.Path = reltorootpath
	page.Type = "shallow"
	if d.IsDir() {
		page.Type = "deep"
	}

	actualabspath := currentpath
	if d.IsDir() {
		actualabspath = filepath.Join(currentpath, "index")
	}

	if req.Mode != "ghost" && req.Meta {
		page.Metadata = f.getMetadata(actualabspath)
	}

	if req.Mode != "ghost" && req.Cont {
		page.Content = f.getContent(actualabspath)
	}

	page.Sorting = req.Sort
	page.Children = nil
	if d.IsDir() {
		page.Children = []*types.Page{}
	}

	page.Stage = "local"
	if req.Mode == "ghost" {
		page.Stage = "ghost"
	}
	page.Og = f.snapPage(page)

	c.Pages[page.Path] = page

	return nil
}

func (f *Filesystem) getMetadata(abspath string) map[string]any {

	metadata := map[string]any{}

	lines := f.getContent(abspath)
	lines, _ = utils.SplitLines(lines)

	for _, line := range lines {
		key, value := utils.SplitTwo(line, ":")
		for k, v := range f.Standards {
			if slices.Contains(v, key) {
				switch k {
				case ".{int}":
					metadata[key], _ = strconv.Atoi(value)
				case ".{yesno}":
					metadata[key] = strings.ToLower(strings.TrimSpace(value)) == "yes"
				case ".{string}":
					metadata[key] = strings.TrimSpace(value)
				case ".{strings}":
					metadata[key] = utils.SplitMore(value, ",")
				default:
					if strings.HasPrefix(k, ".{yyyy.mm.dd") {
						metadata[key] = utils.ParseTime(value)
					}
				}
			}
		}
	}

	pagepath := filepath.Clean(abspath)
	if filepath.Base(pagepath) == "index" {
		pagepath = filepath.Dir(pagepath)
	}
	name := filepath.Base(pagepath)
	if len(name) >= len("2006.01.02") {
		if date, err := time.Parse("2006.01.02", name[:len("2006.01.02")]); err == nil {
			metadata["time"] = date
		}
	}
	if filepath.Base(filepath.Dir(pagepath)) == "proj" && strings.HasPrefix(name, "@") {
		if dollar := strings.LastIndexByte(name, '$'); dollar > 1 {
			if priority, err := strconv.Atoi(name[dollar+1:]); err == nil {
				metadata["priority"] = priority
			}
		}
	}

	return metadata
}

func (f *Filesystem) getContent(abspath string) []string {

	data ,err := os.ReadFile(abspath)
	if err != nil { return []string{} }
	return strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
}

func (f *Filesystem) checkPath(abspath string) string {

	info, err := os.Stat(abspath)
	if err != nil { return "" }
	if info.IsDir() { return "dir" }
	return "file"
}

func (f *Filesystem) standardizePaths(path string) (string, string, string) {

	root := f.whereIsRoot()
	abspath := filepath.Join(root, filepath.Clean("/"+path))
	relpath, err := filepath.Rel(root, abspath)
	if err != nil {
		relpath = "."
	}
	return root, relpath, abspath
}

func (f *Filesystem) whereIsRoot() string {

	if f.root != "" { return f.root }

	exe, err := os.Executable()
	if err != nil { panic(err) }	
	
	if f.api { return filepath.Join(filepath.Dir(exe), "prsnlspc") }

	home, err := os.UserHomeDir()
	return filepath.Join(home, "prsnlspc")

}

func sortChildren(name string, pages []*types.Page) {

	switch name {

	case "normal":
		sort.SliceStable(pages, func(i, j int) bool {
			ideep := pages[i].Type == "deep"
			jdeep := pages[j].Type == "deep"
			if ideep != jdeep {
				return ideep
			}
			return pages[i].Name < pages[j].Name
		})

	case "time":
		sort.SliceStable(pages, func(i, j int) bool {
			itime, _ := pages[i].Metadata["time"].(time.Time)
			jtime, _ := pages[j].Metadata["time"].(time.Time)
			return itime.After(jtime)
		})

	case "depth":
		sort.SliceStable(pages, func(i, j int) bool {
			idepth := utils.PathDepth(pages[i].Path)
			jdepth := utils.PathDepth(pages[j].Path)
			return idepth < jdepth
		})

	case "priority":
		sort.SliceStable(pages, func(i, j int) bool {
			ipriority, _ := pages[i].Metadata["priority"].(int)
			jpriority, _ := pages[j].Metadata["priority"].(int)
			return ipriority > jpriority
		})
	}

	sort.SliceStable(pages, func(i, j int) bool {
		return !strings.HasPrefix(pages[i].Name, ".") && strings.HasPrefix(pages[j].Name, ".")
	})
}

func (f *Filesystem) applyFilters(filters []string, pages map[string]*types.Page) map[string]*types.Page {

	for _, name := range filters {
		pages = f.filter(name, pages)
	}
	
	return pages
}

func (f *Filesystem) filter(name string, pages map[string]*types.Page) map[string]*types.Page {

	one, two := utils.SplitTwo(name, ":")

	switch one {

	case "latest":

		if two == "" { two = "12" }
		list := []*types.Page{}
		for _, p := range pages {
			list = append(list, p)
		}
		sortChildren("time", list)
		newlist := []*types.Page{}
		for i, itm := range list {
			t, _ := strconv.Atoi(two)
			if i == t {
				break
			}
			newlist = append(newlist, itm)
		}
		pages = map[string]*types.Page{}
		for _, itm := range newlist {
			pages[itm.Path] = itm
		}
		return pages

	case "shallow":

		filtered := map[string]*types.Page{}
		for path, p := range pages {
			if p.Type == "shallow" {
				filtered[path] = p
			}
		}
		return filtered

	case "deep":

		filtered := map[string]*types.Page{}
		for path, p := range pages {
			if p.Type == "deep" {
				filtered[path] = p
			}
		}
		return filtered

	default:

		return pages

	}

}

func (f *Filesystem) Sync(req *types.Syncing) chan *types.Patch {
	channel := make(chan *types.Patch, 1)
	var patch *types.Patch
	if req.Api {
		_, patch = f.SyncApi(req)
	} else {
		_, patch = f.SyncLocal(req)
	}
	if patch != nil && len(patch.Commands) > 0 {
		f.Stack = append(f.Stack, patch)
		f.schedulePatch(patch)
		channel <- patch
		if f.Patch != nil {
			go func() { f.Patch <- patch }()
		}
	}
	close(channel)
	return channel
}

func (f *Filesystem) SyncLocal(req *types.Syncing) (bool, *types.Patch) {

	patch := f.newPatch()
	return true, f.walkSync(patch, req.Hard, req.Branch)

}

func (f *Filesystem) schedulePatch(patch *types.Patch) {

	patch.Deadline = time.Now().Add(utils.PatchTime)
	patch.Timer = time.AfterFunc(utils.PatchTime, func() {
		f.ApplyPatch(patch)
	})
}

func (f *Filesystem) ApplyPatch(p *types.Patch) {
	if p == nil { return }
	p.Once.Do(func() {
		if p.Timer != nil { p.Timer.Stop() }
		for _, fn := range p.Commands { fn() }
		if p.Done != nil { close(p.Done) }
	})
}

func (f *Filesystem) CancelPatch(p *types.Patch) {
	if p == nil { return }
	p.Once.Do(func() {
		if p.Timer != nil { p.Timer.Stop() }
		if p.Done != nil { close(p.Done) }
	})
}

func (f *Filesystem) walkSync(patch *types.Patch, hard bool, page *types.Page) *types.Patch {

	f.syncOnce(patch, page)

	if hard {
		for _, child := range page.Children {
			f.walkSync(patch, true, child)
		}
	}

	return patch
}

func (f *Filesystem) syncOnce(patch *types.Patch, page *types.Page) *types.Patch {


	switch page.Stage {
	case "ghost":

	case "draft":

		patch.Commands = append(patch.Commands, func() {

				_, _, abspath := f.standardizePaths(page.Path)

				target := abspath
				if page.Type == "deep" {
					target = filepath.Join(abspath, "index")
				}

				os.MkdirAll(filepath.Dir(target), 0755)
				os.WriteFile(target, []byte(strings.Join(page.Content, "\n")), 0644)

				page.Og = f.snapPage(page)
				page.Stage = "edit"
			})

		patch.Description = append(patch.Description, []string{page.Path, "draft", "edit"})

	case "edit":

		patch.Commands = append(patch.Commands, func() {

				_, _, abspath := f.standardizePaths(page.Path)

				target := abspath
				if page.Type == "deep" {
					target = filepath.Join(abspath, "index")
				}

				os.MkdirAll(filepath.Dir(target), 0755)
				os.WriteFile(target, []byte(strings.Join(page.Content, "\n")), 0644)

				page.Og = f.snapPage(page)
				page.Stage = "local"
			})

		patch.Description = append(patch.Description, []string{page.Path, "edit", "local"})

	case "local", "*api":

			patch.Commands = append(patch.Commands, f.apiSyncCommand(page))
			patch.Description = append(patch.Description, []string{page.Path, page.Stage, "api"})
	case "api":

		patch.Commands = append(patch.Commands, func() {

				_, _, abspath := f.standardizePaths(page.Path)

				target := abspath
				if page.Type == "deep" {
					target = filepath.Join(abspath, "index")
				}

				os.MkdirAll(filepath.Dir(target), 0755)
				os.WriteFile(target, []byte(strings.Join(page.Content, "\n")), 0644)

				page.Og = f.snapPage(page)
				page.Stage = "local"
			})
		patch.Description = append(patch.Description, []string{page.Path, "api", "local"})

	case "doom":

		patch.Commands = append(patch.Commands, func() {
				_, _, abspath := f.standardizePaths(page.Og.Path)

				if page.Type == "deep" {
					os.RemoveAll(abspath)

				} else {
					os.Remove(abspath)
				}
				delete(f.Cache, page.Path)
			})
		patch.Description = append(patch.Description, []string{page.Path, "doom", "x"})

	case "move":

		patch.Commands = append(patch.Commands, func() {

				_, _, oldabspath := f.standardizePaths(page.Og.Path)
				_, _, newabspath := f.standardizePaths(page.Path)

				if page.Type == "deep" {
					if page.Og.Path != page.Path {
						os.MkdirAll(filepath.Dir(newabspath), 0755)
						os.Rename(oldabspath, newabspath)
					}

					target := filepath.Join(newabspath, "index")
					os.MkdirAll(newabspath, 0755)
					os.WriteFile(target, []byte(strings.Join(page.Content, "\n")), 0644)
				} else {
					os.Remove(oldabspath)

					os.MkdirAll(filepath.Dir(newabspath), 0755)
					os.WriteFile(newabspath, []byte(strings.Join(page.Content, "\n")), 0644)
				}

				if page.Og.Path != page.Path {
					delete(f.Cache, page.Og.Path)
				}

				f.Cache[page.Path] = page

				page.Og = f.snapPage(page)
				page.Stage = "local"

			})
		patch.Description = append(patch.Description, []string{page.Path, "move", "local"})

	}

	return patch
}

func (f *Filesystem) SyncApi(req *types.Syncing) (bool, *types.Patch) {


	patch := f.newPatch()

	if req.Branch == nil {
		return false, patch
	}

	return true, f.walkSyncApi(patch, req.Hard, req.Branch)

}

func (f *Filesystem) walkSyncApi(patch *types.Patch, hard bool, page *types.Page) *types.Patch {

	f.syncOnceApi(patch, page)

	if hard {
		for _, child := range page.Children {
			f.walkSyncApi(patch, true, child)
		}
	}

	return patch
}

func (f *Filesystem) syncOnceApi(patch *types.Patch, page *types.Page) *types.Patch {

	switch page.Stage {
	case "ghost":
	case "api":

	case "draft", "edit", "local", "*api", "move":

		patch.Commands = append(patch.Commands, func() {

				oldpath := page.Path
				if page.Og != nil {
					oldpath = page.Og.Path
				}

				body, err := json.Marshal(page)
				if err != nil { page.Stage = "*api"; return }

				httpreq, err := http.NewRequest(http.MethodPost, f.Url+"/sync", bytes.NewReader(body))
				if err != nil { page.Stage = "*api"; return }

				httpreq.Header.Set("Authorization", "Bearer "+f.Token)
				httpreq.Header.Set("Content-Type", "application/json")

				resp, err := http.DefaultClient.Do(httpreq)
				if err != nil { page.Stage = "*api"; return }
				defer resp.Body.Close()

				if resp.StatusCode < 200 || resp.StatusCode >= 300 {
					page.Stage = "*api"
					return
				}

				if oldpath != page.Path {
					delete(f.Cache, oldpath)
				}
				f.Cache[page.Path] = page

				page.Og = f.snapPage(page)
				page.Stage = "api"
			})

		patch.Description = append(patch.Description, []string{page.Path, page.Stage, "api"})

	case "doom":

		patch.Commands = append(patch.Commands, func() {

				body, err := json.Marshal(page)
				if err != nil { page.Stage = "*api"; return }

				httpreq, err := http.NewRequest(http.MethodPost, f.Url+"/sync", bytes.NewReader(body))
				if err != nil { page.Stage = "*api"; return }

				httpreq.Header.Set("Authorization", "Bearer "+f.Token)
				httpreq.Header.Set("Content-Type", "application/json")

				resp, err := http.DefaultClient.Do(httpreq)
				if err != nil { page.Stage = "*api"; return }
				defer resp.Body.Close()

				if resp.StatusCode < 200 || resp.StatusCode >= 300 {
					page.Stage = "*api"
					return
				}

				if page.Og != nil {
					delete(f.Cache, page.Og.Path)
				}
				delete(f.Cache, page.Path)
			})

		patch.Description = append(patch.Description, []string{page.Path, "doom", "x"})

	}

	return patch
}

func (f *Filesystem) apiSyncCommand(page *types.Page) func() {

	return func() {

		body, err := json.Marshal(page)
		if err != nil {
			page.Stage = "*api"
			return
		}

		httpreq, err := http.NewRequest(http.MethodPost, f.Url+"/sync", bytes.NewReader(body))
		if err != nil {
			page.Stage = "*api"
			return
		}
		httpreq.Header.Set("Authorization", "Bearer "+f.Token)
		httpreq.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(httpreq)
		if err != nil {
			page.Stage = "*api"
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			page.Stage = "*api"
			return
		}

		page.Og = f.snapPage(page)
		page.Stage = "api"
	}
}
