package filesystem

import (
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func (f *Filesystem) Load (req *LoadRequest) {
	f.LoadRequest <- req
}

func (f *Filesystem) loadingRoudabout() {

	go func() {
		for req := range f.LoadRequest {

			if req.Api {
				if f.Api == nil { f.Page <- errorPage ; continue }
				newptr, err := f.Api.Fetch(req.Path, req.Depth)
				if err != nil || newptr == nil { f.Page <- errorPage ; continue }
				stampApiTree(newptr)
				filterTree(newptr, parseFilter(req.Filter))
				newptr.Og = snapPage(newptr)
				f.mu.Lock()
				f.merge(newptr)
				result := f.point(newptr)
				f.mu.Unlock()
				f.Page <- result
				continue
			}

			switch req.Mode {

			case "stale":
				f.mu.Lock()
				ptr := findPointer(f.Cache, req.Path)
				f.mu.Unlock()
				if ptr == nil { f.Page <- errorPage ; continue }
				f.Page <- ptr

			case "ghost":
				ok, newptr := f.loadFresh(req)
				if !ok { f.Page <- newptr ; continue }
				f.mu.Lock()
				f.merge(newptr)
				result := f.point(newptr)
				f.mu.Unlock()
				f.Page <- result
				continue

			case "fresh":
				ok, newptr := f.loadFresh(req)
				if !ok { f.Page <- newptr ; continue }
				f.mu.Lock()
				f.merge(newptr)
				result := f.point(newptr)
				f.mu.Unlock()
				f.Page <- result
				continue

			} 
		}
	}()
}

func stampApiTree(p *Page) {
	if p == nil { return }
	p.Stage = "api"
	for _, child := range p.Children {
		stampApiTree(child)
	}
}

func filterTree(p *Page, filter func(*Page) bool) {
	if p == nil || filter == nil { return }
	var kept []*Page
	for _, child := range p.Children {
		if filter(child) {
			filterTree(child, filter)
			kept = append(kept, child)
		}
	}
	p.Children = kept
}

func (f *Filesystem) loadFresh (req *LoadRequest) (ok bool, newptr *Page) {

	root, _, abspath := getPaths(req.Path)
	ok, typee := checkPath(abspath)
	if !ok { return false, errorPage }

	collector := &pageCollector{filter: parseFilter(req.Filter)}
	switch typee {
		case "file":
			info, err := os.Lstat(abspath)
			if err != nil { return false, errorPage }
			
			entry := fs.FileInfoToDirEntry(info)
			err = collector.processEntries(abspath, entry, nil, root, abspath, req)
			if err != nil { return false, errorPage }

		// there may be more entries ..
		case "dir":
			err := filepath.WalkDir(abspath, func(path string, d fs.DirEntry, err error) error {
				return collector.processEntries(path, d, err, root, abspath, req)
			})
			if err != nil { return false, errorPage }
	}

	if len(collector.Pages) == 0 { return false, errorPage }

	if typee == "dir" {
		
		nested := nestPage(collector.Pages)
		newptr = findPointer(nested, req.Path)
		if newptr == nil { return false, errorPage }

		newptr = sortPage(newptr, req.Sort)
		newptr.Og = snapPage(newptr)
		return true, newptr
	}


	for _, page := range collector.Pages {
		if page.Path == req.Path {
			newptr = page
			break
		}
	}
	if newptr == nil { return false, errorPage }

	newptr = sortPage(newptr, req.Sort)
	newptr.Og = snapPage(newptr)
	return true, newptr
}

type pageCollector struct {
	Pages  []*Page
	filter func(*Page) bool
}
func (c *pageCollector) processEntries (currentpath string, d fs.DirEntry, err error, root string, startabspath string, req *LoadRequest) error {

	if err != nil { return err }
	if d == nil { return nil }

	if currentpath != startabspath && (strings.HasPrefix(d.Name(), ".") || d.Name() == "index") {
		if d.IsDir() { return filepath.SkipDir }
		return nil
	}

	reltostartpath, _ := filepath.Rel(startabspath, currentpath)
	reltorootpath, _ := filepath.Rel(root, currentpath)

	depth := calculatePathDepth(reltostartpath)
	if req.Depth != -1 && depth > req.Depth {
		if d.IsDir() { return filepath.SkipDir }
		return nil
	}

	page := newPage()
	if req.Mode == "ghost" {
		page.Stage = "ghost"
	} else {
		page.Stage = "local"
	}

	page.Name = d.Name()
	page.Path = reltorootpath
	page.Type = "shallow"
	if d.IsDir() { page.Type = "deep" }

	actualabspath := currentpath
	if d.IsDir() { actualabspath = filepath.Join(currentpath, "index") }

	if req.Mode != "ghost" && req.Opts { page.Options = getOptions(actualabspath)}
	if req.Mode != "ghost" && req.Cont { page.Content = getContent(actualabspath)}
	if req.Mode != "ghost" && req.Meta { page.Metadata = getMetadata(actualabspath)}

	page.Sorting = req.Sort

	if req.Mode != "ghost" {
		page.Og = snapPage(page)
	}

	if c.filter != nil && !c.filter(page) {
		if d.IsDir() { return filepath.SkipDir }
		return nil
	}

	c.Pages = append(c.Pages, page)
	return nil
}

func getOptions (abspath string) map[string]any {

	root, _, _ := getPaths(abspath)
	ok, typee := checkPath(abspath)
	if !ok { return map[string]any{} }

	dir := abspath
	if typee == "file" { dir = filepath.Dir(abspath) }

	for {
		optspath := filepath.Join(dir, ".options")
		
		if root == filepath.Clean(dir) { return processHeader(optspath) }

		ok, typee := checkPath(optspath)
		if ok && typee == "file" { return processHeader(optspath) }

		dir = filepath.Dir(dir)
	}
}

func getMetadata (abspath string) map[string]any {

	return processHeader(abspath)
}

func processHeader(abspath string) map[string]any {
	return ParseMetadataFromContent(getContent(abspath), filepath.Base(abspath))
}

func ParseMetadataFromContent(lines []string, name string) map[string]any {

	baseopts := getBaseOptions()
	lines, _ = splitHeader(lines)

	final := map[string]any{}

	for _, line := range lines {

		key, value := splitIntoTwo(line, ":")
		key = strings.ToLower(key)
		keytype := getKeytype(baseopts, key)

		switch keytype {
		case "time":
			_, final[key] = parseTime(value)
			final["time"] = final[key]
		case "int":
			final[key], _ = strconv.Atoi(value)
		case "yesno":
			final[key] = strings.ToLower(trimString(value)) == "yes"
		case "string":
			final[key] = trimString(value)
		case "strings":
			final[key] = trimStrings(splitIntoMore(value, ","))
		default:
			final[key] = value
		}
	}

	typee, t := parseTime(name)
	if typee == "date" { final["time"] = t }

	return final
}

func getBaseOptions () map[string][]string {

	_, _, abspath := getPaths(".options")

	lines := getContent(abspath)
	_, lines = splitHeader(lines)

	final := map[string][]string{}

	for _, line := range lines {

		key, value := splitIntoTwo(line, ":")
		values := splitIntoMore(value, ",")

		if final[key] == nil {
			final[key] = values
			continue
		}
		final[key] = append(final[key], values...)
	}

	return final
}

func getKeytype (baseopts map[string][]string, key string) string {

	for keyy, values := range baseopts {
		for _, value := range values {
			if key == value { return keyy }
		}
	}
	return ""
}

func getContent (abspath string) []string {

	data, err := os.ReadFile(abspath)
	if err != nil { return []string{} }
	return strings.Split(string(data), "\n")
}
