package picker

import (
	"env/apiclient"
	"env/cli/menu"
	"env/filesystem"
	"env/routines"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const pickerSyncDelay = 5 * time.Second

func (p *Picker) Input(newinput *routines.Input) {

	switch newinput.Key {

	case "up":
		p.moveVertical(1)

	case "down":
		p.moveVertical(-1)

	case "ctrl+up":
		p.scrollPreview(-1)

	case "ctrl+down":
		p.scrollPreview(1)

	case "ctrl+p":
		p.cyclePanelMode()

	case "char":
		p.Prompt += string(newinput.Char)
		p.Selected = 0
		p.PreviewOffset = 0
		p.updateItems()

	case "backspace":
		p.backspacePrompt()

	case "enter":
		p.enter()

	case "ctrl+enter":
		p.openSelectedLauncherActions()

	case "ctrl+backspace":
		if strings.TrimSpace(p.Prompt) == "" {
			p.cycleSorting()
		} else {
			p.cycleMode()
		}

	case "ctrl+right":
		p.cycleScope()
	}
}

func (p *Picker) cyclePanelMode() {
	switch p.PanelMode {
	case "list":
		p.PanelMode = "preview"
	case "preview":
		p.PanelMode = "both"
	default:
		p.PanelMode = "list"
	}
}

func (p *Picker) backspacePrompt() {
	if p.Prompt == "" {
		return
	}

	runes := []rune(p.Prompt)
	if len(runes) == 0 {
		return
	}

	p.Prompt = string(runes[:len(runes)-1])
	p.Selected = 0
	p.PreviewOffset = 0
	p.updateItems()
}

func (p *Picker) enter() {
	item := p.selectedItem()
	if item == nil {
		return
	}

	if item.Kind == "parent" {
		p.enterParentContext()
		return
	}

	if item.Page == nil {
		return
	}

	page := item.Page

	if page.Type != "deep" {
		return
	}

	p.changeContext(page.Path)
}

func (p *Picker) openSelectedLauncherActions() {
	if p.Parent == nil {
		return
	}

	items := []menu.Item{}

	item := p.selectedItem()

	if item != nil && item.Kind == "parent" {
		items = append(items, menu.Item{
			Name: "new draft here",
			Pin:  true,
			Run:  func(prompt string) { p.createDraftInContext(prompt, "shallow") },
		})

		switch p.Path {
		case "d.fami":
			items = append(items, menu.Item{
				Name: "new family",
				Pin:  true,
				Run:  func(prompt string) { p.createDraftInContext(prompt, "deep") },
			})

		case "e.proj":
			items = append(items, menu.Item{
				Name: "new project",
				Pin:  true,
				Run:  func(prompt string) { p.createDraftInContext(prompt, "deep") },
			})
		}
	}

	if item != nil && item.Page != nil {
		items = append(items, p.selectedPageLauncherItems(item.Page)...)
	}

	if len(items) == 0 {
		return
	}

	p.Parent.OpenLauncher("PICKER ACTIONS", items)
}

func (p *Picker) selectedPageLauncherItems(page *filesystem.Page) []menu.Item {
	if page == nil {
		return nil
	}

	items := []menu.Item{}

	isBranchable := page.Type == "deep" && len(page.Children) > 0
	isProtected := isPickerProtectedPath(page.Path)
	stage := page.Stage
	if stage == "" {
		stage = pageStageLocal
	}
	canReadBranch := page.Type == "deep" && (len(page.Children) > 0 || stage == pageStageGhost || stage == pageStageApi)

	addOpenEditor := func() {
		items = append(items, menu.Item{
			Name:    "open in editor",
			Command: "open in editor",
			Run:     func(_ string) { p.Parent.AddClients("editor:"+page.Path, "next") },
		})
	}
	addReadFresh := func() {
		items = append(items, menu.Item{
			Name:    "read fresh",
			Command: "read fresh",
			Run:     func(_ string) { p.loadFresh(page) },
		})
		if canReadBranch {
			items = append(items, menu.Item{
				Name:    "read fresh branch",
				Command: "read fresh branch",
				Run:     func(_ string) { p.loadFreshDeep(page) },
			})
		}
	}
	addReadAPI := func() {
		if !p.pickerApiAvailable() {
			return
		}
		items = append(items, menu.Item{
			Name:    "read api",
			Command: "read api",
			Run:     func(_ string) { p.loadAPI(page) },
		})
		if canReadBranch {
			items = append(items, menu.Item{
				Name:    "read api branch",
				Command: "read api branch",
				Run:     func(_ string) { p.loadAPIDeep(page) },
			})
		}
	}
	addSyncLocal := func() {
		items = append(items, menu.Item{
			Name:    "synclocal",
			Command: "synclocal",
			Run:     func(_ string) { p.syncSinglePage(page) },
		})
		if isBranchable {
			items = append(items, menu.Item{
				Name:    "synclocal branch",
				Command: "synclocal branch",
				Run:     func(_ string) { p.syncBranchPage(page) },
			})
		}
	}
	addSyncAPI := func() {
		if !p.pickerApiAvailable() {
			return
		}
		items = append(items, menu.Item{
			Name:    "apisync",
			Command: "apisync",
			Run:     func(_ string) { p.syncSinglePageToAPI(page) },
		})
		if isBranchable {
			items = append(items, menu.Item{
				Name:    "apisync branch",
				Command: "apisync branch",
				Run:     func(_ string) { p.syncBranchPageToAPI(page) },
			})
		}
	}
	addDoom := func() {
		if isProtected {
			return
		}
		items = append(items, menu.Item{
			Name:    "doom",
			Command: "doom",
			Run:     func(_ string) { p.doomSinglePage(page) },
		})
		if isBranchable {
			items = append(items, menu.Item{
				Name:    "doom branch",
				Command: "doom branch",
				Run:     func(_ string) { p.doomBranchPage(page) },
			})
		}
	}

	switch stage {
	case pageStageGhost:
		addReadFresh()
		addReadAPI()

	case pageStageLocal:
		addOpenEditor()
		addSyncAPI()
		addDoom()

	case pageStageApi, pageStageEdit, pageStageDraft:
		addOpenEditor()
		if stage == pageStageApi {
			addReadAPI()
		}
		addSyncLocal()
		addDoom()

	case pageStageDoom:
		items = append(items, menu.Item{
			Name:    "undoom",
			Command: "undoom",
			Run:     func(_ string) { p.undoDoomPage(page) },
		})
		addSyncLocal()

	case pageStageConflict:
		addOpenEditor()
		addSyncLocal()
		addSyncAPI()

	default:
		addOpenEditor()
		addDoom()
	}

	if !isProtected {
		items = append(items, menu.Item{
			Name:    "toggle type",
			Command: "toggle type",
			Run:     func(_ string) { p.togglePageType(page) },
		})
	}

	return items
}

func (p *Picker) syncSinglePage(page *filesystem.Page) {
	if p.Parent == nil || page == nil {
		return
	}

	fs := p.Parent.GetFilesystem()
	if fs == nil {
		return
	}

	label := "synclocal " + page.Path
	if page.Stage == pageStageDoom {
		label = "delete " + page.Path
	}
	commitPage := clonePickerPageTree(page)

	p.queuePickerAction(label, func() {
		p.optimisticSyncPage(page, false)
	}, func() {
		req := filesystem.NewSyncRequest()
		req.Branch = commitPage
		req.PageOnly = true
		fs.Sync(req)
	})
}

func (p *Picker) syncBranchPage(page *filesystem.Page) {
	if p.Parent == nil || page == nil {
		return
	}

	commitPage := clonePickerPageTree(page)
	p.queuePickerAction("synclocal branch: "+page.Path, func() {
		p.optimisticSyncBranch(page, false)
	}, func() {
		fs := p.Parent.GetFilesystem()
		if fs == nil {
			return
		}
		req := filesystem.NewSyncRequest()
		req.Branch = commitPage
		fs.Sync(req)
	})
}

func (p *Picker) syncSinglePageToAPI(page *filesystem.Page) {
	if p.Parent == nil || page == nil {
		return
	}

	fs := p.Parent.GetFilesystem()
	if fs == nil {
		return
	}

	commitPage := clonePickerPageTree(page)
	p.queuePickerAction("apisync "+page.Path, func() {
		p.optimisticSyncPage(page, true)
	}, func() {
		req := filesystem.NewSyncRequest()
		req.Api = true
		req.Branch = commitPage
		req.PageOnly = true
		fs.Sync(req)
	})
}

func (p *Picker) syncBranchPageToAPI(page *filesystem.Page) {
	if p.Parent == nil || page == nil {
		return
	}

	commitPage := clonePickerPageTree(page)
	p.queuePickerAction("apisync branch: "+page.Path, func() {
		p.optimisticSyncBranch(page, true)
	}, func() {
		fs := p.Parent.GetFilesystem()
		if fs == nil {
			return
		}
		req := filesystem.NewSyncRequest()
		req.Api = true
		req.Branch = commitPage
		fs.Sync(req)
	})
}

func (p *Picker) pickerApiAvailable() bool {
	if p == nil || p.Parent == nil {
		return false
	}

	sc := p.Parent.GetSyncClient()
	return sc != nil && sc.URL() != ""
}

func (p *Picker) undoDoomPage(page *filesystem.Page) {
	if p.Parent == nil || page == nil {
		return
	}
	page.Stage = ""
	p.StartPath = page.Path
	p.updateItems()
	p.loadFresh(page)
}

func (p *Picker) doomSinglePage(page *filesystem.Page) {
	if p.Parent == nil || page == nil {
		return
	}

	fs := p.Parent.GetFilesystem()
	if fs == nil {
		return
	}

	p.queuePickerAction("doom "+page.Path, func() {
		fs.Doom(page, false)
		p.StartPath = page.Path
		p.updateItems()
	}, nil)
}

func (p *Picker) doomBranchPage(page *filesystem.Page) {
	if p.Parent == nil || page == nil {
		return
	}

	fs := p.Parent.GetFilesystem()
	if fs == nil {
		return
	}

	p.queuePickerAction("doom branch: "+page.Path, func() {
		doomTree(fs, page)
		p.StartPath = page.Path
		p.updateItems()
	}, nil)
}

func doomTree(fs *filesystem.Filesystem, page *filesystem.Page) {
	if page == nil {
		return
	}
	fs.Doom(page, false)
	for _, child := range page.Children {
		doomTree(fs, child)
	}
}

func (p *Picker) resolveConflict(page *filesystem.Page, sc *apiclient.Client) {
	if page == nil || sc == nil || p.Parent == nil {
		return
	}

	branch, relPath := splitBranchPath(page.Path)
	if branch == "" || relPath == "" {
		return
	}

	serverHash := sc.ConflictHash(page.Path)
	if serverHash == "" {
		return
	}

	fs := p.Parent.GetFilesystem()
	if fs == nil {
		return
	}

	content := strings.Join(page.Content, "\n")
	sc.ClearConflictHash(page.Path)

	go func() {
		resp, err := sc.Write(apiclient.WriteRequest{
			Branch: branch,
			Filter: "",
			Files:  []apiclient.FileEntry{{Path: relPath, Content: content, BaseHash: serverHash}},
		})
		if err != nil {
			sc.SetConflictHash(page.Path, serverHash)
			return
		}

		for _, conflict := range resp.Conflicts {
			if conflict.Path == relPath {
				// servidor mudou de novo durante a resolução
				sc.SetConflictHash(page.Path, extractServerHash(conflict.ConflictContent))
				conflictPage := &filesystem.Page{
					Name:    page.Name,
					Path:    page.Path,
					Type:    page.Type,
					Content: strings.Split(conflict.ConflictContent, "\n"),
					Stage:   pageStageConflict,
				}
				fs.EditContent(conflictPage, conflictPage.Content)
				syncReq := filesystem.NewSyncRequest()
				syncReq.Branch = conflictPage
				fs.Sync(syncReq)
				return
			}
		}

		for _, written := range resp.Written {
			if written == relPath {
				sc.SetBaseHash(page.Path, apiclient.HashContent(content))
				return
			}
		}
	}()
}

func (p *Picker) pushFolderToAPI(page *filesystem.Page, sc *apiclient.Client) {
	if page == nil || sc == nil || p.Parent == nil {
		return
	}

	branch, folderRel := splitBranchPath(page.Path)
	if branch == "" {
		return
	}

	fs := p.Parent.GetFilesystem()
	if fs == nil {
		return
	}

	root := pickerRootPath()
	branchDir := filepath.Join(root, branch)
	folderDir := filepath.Join(root, page.Path)

	go func() {
		var entries []apiclient.FileEntry

		filepath.Walk(folderDir, func(fpath string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || strings.HasSuffix(fpath, ".conflict") {
				return nil
			}
			rel, _ := filepath.Rel(branchDir, fpath)
			rel = filepath.ToSlash(rel)
			data, err := os.ReadFile(fpath)
			if err != nil {
				return nil
			}
			fullPath := branch + "/" + rel
			entries = append(entries, apiclient.FileEntry{
				Path:     rel,
				Content:  string(data),
				BaseHash: sc.BaseHash(fullPath),
			})
			return nil
		})

		filter := ""
		if folderRel == "" {
			filter = "*"
		}

		resp, err := sc.Write(apiclient.WriteRequest{
			Branch: branch,
			Filter: filter,
			Files:  entries,
		})
		if err != nil {
			return
		}

		contentByRel := map[string]string{}
		for _, e := range entries {
			contentByRel[e.Path] = e.Content
		}

		for _, conflict := range resp.Conflicts {
			fullPath := branch + "/" + conflict.Path
			sc.SetConflictHash(fullPath, extractServerHash(conflict.ConflictContent))
			conflictPage := &filesystem.Page{
				Path:    fullPath,
				Type:    "shallow",
				Content: strings.Split(conflict.ConflictContent, "\n"),
				Stage:   pageStageConflict,
			}
			fs.EditContent(conflictPage, conflictPage.Content)
			syncReq := filesystem.NewSyncRequest()
			syncReq.Branch = conflictPage
			fs.Sync(syncReq)
		}

		for _, written := range resp.Written {
			fullPath := branch + "/" + written
			if content, ok := contentByRel[written]; ok {
				sc.SetBaseHash(fullPath, apiclient.HashContent(content))
			}
		}
	}()
}

func (p *Picker) pullFolderFromAPI(page *filesystem.Page, sc *apiclient.Client) {
	if page == nil || sc == nil || p.Parent == nil {
		return
	}

	branch, folderRel := splitBranchPath(page.Path)
	if branch == "" {
		return
	}

	fs := p.Parent.GetFilesystem()
	if fs == nil {
		return
	}

	go func() {
		resp, err := sc.Read(apiclient.ReadRequest{Branch: branch, Filter: ""})
		if err != nil {
			return
		}

		conflictSet := map[string]bool{}
		for _, c := range resp.Conflicts {
			conflictSet[c] = true
		}

		for _, file := range resp.Files {
			if folderRel != "" && !strings.HasPrefix(file.Path, folderRel+"/") {
				continue
			}

			fullPath := branch + "/" + file.Path
			isConflicted := conflictSet[file.Path]

			stage := pageStageEdit
			if isConflicted {
				stage = pageStageConflict
			}

			filePage := &filesystem.Page{
				Path:    fullPath,
				Type:    "shallow",
				Content: strings.Split(file.Content, "\n"),
				Stage:   stage,
			}
			fs.EditContent(filePage, filePage.Content)
			syncReq := filesystem.NewSyncRequest()
			syncReq.Branch = filePage
			fs.Sync(syncReq)

			if !isConflicted {
				sc.SetBaseHash(fullPath, apiclient.HashContent(file.Content))
			}
		}
	}()
}

func extractServerHash(conflictContent string) string {
	for _, line := range strings.SplitN(conflictContent, "\n", 10) {
		if strings.HasPrefix(line, "server-hash: ") {
			return strings.TrimPrefix(line, "server-hash: ")
		}
	}
	return ""
}

func pickerRootPath() string {
	if root := strings.TrimSpace(os.Getenv("ENV_ROOT")); root != "" {
		return filepath.Clean(root)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "/home/asdf/prsnl.spc"
	}
	return filepath.Join(home, "prsnl.spc")
}

func splitBranchPath(path string) (branch, relPath string) {
	path = filepath.ToSlash(filepath.Clean(path))
	i := strings.Index(path, "/")
	if i < 0 {
		return path, ""
	}
	return path[:i], path[i+1:]
}

func (p *Picker) pushFileToAPI(page *filesystem.Page, sc *apiclient.Client) {
	if page == nil || sc == nil || p.Parent == nil {
		return
	}

	branch, relPath := splitBranchPath(page.Path)
	if branch == "" || relPath == "" {
		return
	}

	content := strings.Join(page.Content, "\n")
	baseHash := sc.BaseHash(page.Path)
	fs := p.Parent.GetFilesystem()
	if fs == nil {
		return
	}

	go func() {
		resp, err := sc.Write(apiclient.WriteRequest{
			Branch: branch,
			Filter: "",
			Files:  []apiclient.FileEntry{{Path: relPath, Content: content, BaseHash: baseHash}},
		})
		if err != nil {
			return
		}

		for _, conflict := range resp.Conflicts {
			if conflict.Path == relPath {
				sc.SetConflictHash(page.Path, extractServerHash(conflict.ConflictContent))
				conflictPage := &filesystem.Page{
					Name:    page.Name,
					Path:    page.Path,
					Type:    page.Type,
					Content: strings.Split(conflict.ConflictContent, "\n"),
					Stage:   pageStageConflict,
				}
				fs.EditContent(conflictPage, conflictPage.Content)
				syncReq := filesystem.NewSyncRequest()
				syncReq.Branch = conflictPage
				fs.Sync(syncReq)
				return
			}
		}

		for _, written := range resp.Written {
			if written == relPath {
				sc.SetBaseHash(page.Path, apiclient.HashContent(content))
				return
			}
		}
	}()
}

func (p *Picker) pullFileFromAPI(page *filesystem.Page, sc *apiclient.Client) {
	if page == nil || sc == nil || p.Parent == nil {
		return
	}

	branch, relPath := splitBranchPath(page.Path)
	if branch == "" || relPath == "" {
		return
	}

	fs := p.Parent.GetFilesystem()
	if fs == nil {
		return
	}

	go func() {
		resp, err := sc.Read(apiclient.ReadRequest{Branch: branch, Filter: relPath})
		if err != nil {
			return
		}

		for _, file := range resp.Files {
			if file.Path != relPath {
				continue
			}

			isConflicted := false
			for _, c := range resp.Conflicts {
				if c == relPath {
					isConflicted = true
					break
				}
			}

			stage := pageStageEdit
			if isConflicted {
				stage = pageStageConflict
			}

			filePage := &filesystem.Page{
				Name:    page.Name,
				Path:    page.Path,
				Type:    page.Type,
				Content: strings.Split(file.Content, "\n"),
				Stage:   stage,
			}
			fs.EditContent(filePage, filePage.Content)
			syncReq := filesystem.NewSyncRequest()
			syncReq.Branch = filePage
			fs.Sync(syncReq)

			if !isConflicted {
				sc.SetBaseHash(page.Path, apiclient.HashContent(file.Content))
			}
			return
		}
	}()
}

func (p *Picker) selectedItem() *Match {
	if len(p.Items) == 0 {
		return nil
	}

	if p.Selected < 0 || p.Selected >= len(p.Items) {
		return nil
	}

	return p.Items[p.Selected]
}

func (p *Picker) enterParentContext() {
	if p.isRootContext() {
		return
	}

	parent := parentPathOf(p.Path)
	p.changeContext(parent)
}

func (p *Picker) changeContext(path string) {
	if path == "" {
		path = "."
	}

	p.Path = path
	p.StartPath = ""
	p.Prompt = ""
	p.Selected = 0
	p.PreviewOffset = 0
	p.updateItems()
}

func (p *Picker) readSkeletonFresh(page *filesystem.Page) {
	p.loadFreshDeep(page)
}

func (p *Picker) loadFresh(page *filesystem.Page) {
	if p.Parent == nil || page == nil {
		return
	}

	fs := p.Parent.GetFilesystem()
	if fs == nil {
		return
	}

	req := filesystem.NewLoadRequest()
	req.Mode = "fresh"
	req.Path = page.Path
	req.Depth = 0
	req.Opts = true
	req.Cont = true
	req.Meta = true

	p.queuePickerAction("read fresh "+page.Path, func() {
		fs.Load(req)
	}, nil)
}

func (p *Picker) loadFreshDeep(page *filesystem.Page) {
	if p.Parent == nil || page == nil {
		return
	}

	fs := p.Parent.GetFilesystem()
	if fs == nil {
		return
	}

	req := filesystem.NewLoadRequest()
	req.Mode = "fresh"
	req.Path = page.Path
	req.Depth = -1
	req.Opts = true
	req.Cont = true
	req.Meta = true

	p.queuePickerAction("read fresh branch: "+page.Path, func() {
		fs.Load(req)
	}, nil)
}

func (p *Picker) loadAPI(page *filesystem.Page) {
	p.loadAPIWithDepth(page, 0)
}

func (p *Picker) loadAPIDeep(page *filesystem.Page) {
	p.loadAPIWithDepth(page, -1)
}

func (p *Picker) loadAPIWithDepth(page *filesystem.Page, depth int) {
	if p.Parent == nil || page == nil || !p.pickerApiAvailable() {
		return
	}

	fs := p.Parent.GetFilesystem()
	if fs == nil {
		return
	}

	req := filesystem.NewLoadRequest()
	req.Mode = "fresh"
	req.Api = true
	req.Path = page.Path
	req.Depth = depth
	req.Opts = true
	req.Cont = true
	req.Meta = true

	label := "read api " + page.Path
	if depth != 0 {
		label = "read api branch: " + page.Path
	}

	p.queuePickerAction(label, func() {
		fs.Load(req)
	}, nil)
}

func (p *Picker) queuePickerAction(label string, optimistic func(), commit func()) {
	if p.Parent == nil {
		if optimistic != nil {
			optimistic()
		}
		if commit != nil {
			commit()
		}
		return
	}

	fs := p.Parent.GetFilesystem()
	snapshot := (*filesystem.Page)(nil)
	if fs != nil {
		snapshot = clonePickerPageTree(fs.Cache)
	}

	if optimistic != nil {
		optimistic()
	}

	p.Parent.QueueDelayedAction(label, pickerSyncDelay, func() {
		if commit != nil {
			commit()
		}
	}, func() {
		if fs == nil || snapshot == nil {
			return
		}
		fs.Cache = snapshot
		p.Cache = snapshot
		p.Selected = 0
		p.PreviewOffset = 0
		p.updateItems()
	})
}

func (p *Picker) optimisticSyncPage(page *filesystem.Page, api bool) {
	if page == nil {
		return
	}

	if page.Stage == pageStageDoom {
		p.removePageFromVisibleCache(page.Path)
		p.updateItems()
		return
	}

	if api {
		page.Stage = pageStageApi
	} else {
		page.Stage = pageStageLocal
	}
	page.Diff = nil
	page.Og = clonePickerPageTree(page)
	p.StartPath = page.Path
	p.updateItems()
}

func (p *Picker) optimisticSyncBranch(page *filesystem.Page, api bool) {
	if page == nil {
		return
	}

	if page.Stage == pageStageDoom {
		p.removePageFromVisibleCache(page.Path)
		p.updateItems()
		return
	}

	p.optimisticSyncBranchWalk(page, api)
	p.StartPath = page.Path
	p.updateItems()
}

func (p *Picker) optimisticSyncBranchWalk(page *filesystem.Page, api bool) {
	if page == nil {
		return
	}

	for _, child := range page.Children {
		p.optimisticSyncBranchWalk(child, api)
	}

	switch page.Stage {
	case pageStageDoom:
		p.removePageFromVisibleCache(page.Path)
	case pageStageDraft, pageStageEdit, pageStageConflict, pageStageApi, pageStageLocal:
		if api {
			page.Stage = pageStageApi
		} else {
			page.Stage = pageStageLocal
		}
		page.Diff = nil
		page.Og = clonePickerPageTree(page)
	}
}

func (p *Picker) removePageFromVisibleCache(path string) {
	if p.Cache == nil || path == "" || path == "." {
		return
	}

	p.removePageFromParent(p.Cache, path)
}

func (p *Picker) removePageFromParent(parent *filesystem.Page, path string) bool {
	if parent == nil {
		return false
	}

	for i, child := range parent.Children {
		if child == nil {
			continue
		}
		if child.Path == path {
			parent.Children = append(parent.Children[:i], parent.Children[i+1:]...)
			return true
		}
		if p.removePageFromParent(child, path) {
			return true
		}
	}

	return false
}

func clonePickerPageTree(page *filesystem.Page) *filesystem.Page {
	if page == nil {
		return nil
	}

	clone := &filesystem.Page{
		Name:     page.Name,
		Path:     page.Path,
		Type:     page.Type,
		Options:  clonePickerMap(page.Options),
		Content:  append([]string(nil), page.Content...),
		Metadata: clonePickerMap(page.Metadata),
		Sorting:  page.Sorting,
		Og:       clonePickerPageTree(page.Og),
		Stage:    page.Stage,
		Diff:     append([]string(nil), page.Diff...),
	}

	for _, child := range page.Children {
		clone.Children = append(clone.Children, clonePickerPageTree(child))
	}

	return clone
}

func clonePickerMap(in map[string]any) map[string]any {
	if in == nil {
		return nil
	}

	out := map[string]any{}
	for key, value := range in {
		out[key] = value
	}
	return out
}

func (p *Picker) moveVertical(delta int) {
	if len(p.Items) == 0 {
		p.Selected = 0
		p.PreviewOffset = 0
		return
	}

	newy := p.Selected + delta
	newy = clamp(newy, 0, len(p.Items)-1)
	if newy != p.Selected {
		p.Selected = newy
		p.PreviewOffset = 0
	}
}

func (p *Picker) scrollPreview(delta int) {
	p.PreviewOffset += delta
	if p.PreviewOffset < 0 {
		p.PreviewOffset = 0
	}
}

func clamp(num int, min int, max int) int {
	if max < min {
		return min
	}

	if num < min {
		return min
	}

	if num > max {
		return max
	}

	return num
}

func (p *Picker) cycleMode() {
	switch p.Mode {
	case "literal":
		p.Mode = "fuzzy"

	case "fuzzy":
		p.Mode = "metadata"

	case "metadata":
		p.Mode = "recent"

	case "recent":
		p.Mode = "literal"

	default:
		p.Mode = "literal"
	}

	p.Selected = 0
	p.PreviewOffset = 0
	p.updateItems()
}

func (p *Picker) cycleSorting() {
	switch p.Sorting {
	case "auto":
		p.Sorting = "basic"

	case "basic":
		p.Sorting = "time"

	case "time":
		p.Sorting = "priority"

	case "priority":
		p.Sorting = "auto"

	default:
		p.Sorting = "auto"
	}

	p.Selected = 0
	p.PreviewOffset = 0
	p.updateItems()
}

func (p *Picker) cycleScope() {
	switch p.Scope {
	case "context":
		p.Scope = "tree"
	default:
		p.Scope = "context"
	}

	p.Selected = 0
	p.PreviewOffset = 0
	p.updateItems()
}

func (p *Picker) hydrateDeepFull(page *filesystem.Page) {
	if page == nil || page.Type != "deep" || p.Parent == nil {
		return
	}

	fs := p.Parent.GetFilesystem()
	if fs == nil {
		return
	}

	p.StartPath = page.Path
	p.refreshPage(page)
}

func (p *Picker) refreshPage(page *filesystem.Page) {
	if page == nil || p.Parent == nil {
		return
	}

	fs := p.Parent.GetFilesystem()
	if fs == nil {
		return
	}

	req := filesystem.NewLoadRequest()
	req.Mode = "fresh"
	req.Path = page.Path
	req.Opts = true
	req.Cont = true
	req.Meta = true
	req.Depth = -1
	req.Sort = p.Sorting

	fs.Load(req)
}

func isPickerProtectedPath(path string) bool {
	path = filepath.Clean(path)

	protected := map[string]bool{
		".":         true,
		"prsnl.spc": true,

		"a.log":      true,
		"b.rec":      true,
		"c.rand":     true,
		"d.fami":     true,
		"e.proj":     true,
		".resources": true,
	}

	return protected[path]
}

func (p *Picker) createDraftInContext(name string, typee string) {
	if p.Parent == nil {
		return
	}

	fs := p.Parent.GetFilesystem()
	if fs == nil {
		return
	}

	name = strings.TrimSpace(name)
	if name == "" {
		name = "untitled"
	}
	if typee != "deep" {
		typee = "shallow"
	}

	base := p.Path
	if base == "" || base == "." {
		base = "c.rand"
	}

	path := filepath.Join(base, name)

	if p.findPageByPath(p.Cache, path) != nil {
		return
	}

	_, draft := fs.NewDraft(path)
	if draft == nil {
		return
	}

	draft.Type = typee
	draft.Content = []string{""}
	draft.Stage = pageStageDraft
	if typee == "deep" {
		draft.Children = []*filesystem.Page{}
	}

	p.StartPath = draft.Path
	p.Prompt = ""
	p.updateItems()

	p.Parent.AddClients("editor:"+draft.Path, "next")
}

func (p *Picker) togglePageType(page *filesystem.Page) {
	if p.Parent == nil {
		return
	}

	fs := p.Parent.GetFilesystem()
	if fs == nil || page == nil || isPickerProtectedPath(page.Path) {
		return
	}

	nextType := "deep"
	if page.Type == "deep" {
		nextType = "shallow"
	}

	fs.EditType(page, nextType)
	p.StartPath = page.Path
	p.updateItems()
}
