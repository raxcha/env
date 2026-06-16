package filesystem

import "path/filepath"

func (f *Filesystem) NewDraft(path string) (bool, *Page) {

	p := newPage()
	p.Name = filepath.Base(path)
	p.Path = path
	p.Stage = "draft"

	f.mu.Lock()
	inserted := insertPointer(f.Cache, p)
	parent := findParent(f.Cache, inserted.Path)
	f.mu.Unlock()

	if parent != nil {
		parent = sortPage(parent, parent.Sorting)
	}

	return true, inserted
}

func (f *Filesystem) EditContent(p *Page, newcontent []string) (bool, *Page) {

	ensureOriginalContent(p)
	p.Content = newcontent
	p.Stage = "edited"
	UpdateDiff(p)
	return true, p
}

func (f *Filesystem) EditName(p *Page, newname string) (bool, *Page) {

	if calculatePathDepth(p.Path) == 0 {
		return false, p
	}

	p.Name = newname
	p.Stage = "edited"

	f.mu.Lock()
	parent := findParent(f.Cache, p.Path)
	f.mu.Unlock()
	parent = sortPage(parent, parent.Sorting)
	UpdateDiff(p)
	return true, p
}

func (f *Filesystem) EditPath(p *Page, newpath string) (bool, *Page) {

	oldpath := p.Path
	p.Name = filepath.Base(newpath)
	RepathPage(p, newpath)

	f.mu.Lock()
	removePage(f.Cache, oldpath)
	newnode := insertPointer(f.Cache, p)
	parent := findParent(f.Cache, newnode.Path)
	f.mu.Unlock()

	parent = sortPage(parent, parent.Sorting)
	newnode.Stage = "edited"
	UpdateDiff(newnode)

	return true, newnode
}

func (f *Filesystem) EditType(p *Page, typee string) (bool, *Page) {

	if p.Type == typee {
		return false, p
	}

	if typee == "deep" {
		p.Type = "deep"
		p.Children = []*Page{}
		p.Stage = "edited"
		UpdateDiff(p)
		return true, p
	}

	// f.PromoteChildren(p)
	p.Type = typee
	p.Children = nil
	p.Stage = "edited"
	UpdateDiff(p)
	return true, p
}

func (f *Filesystem) Resort(p *Page, spec string) (bool, *Page) {

	sortPage(p, spec)
	return true, p
}

func (f *Filesystem) EditMetadata(p *Page, newlines []string) (bool, *Page) {

	header, _ := splitHeader(p.Content)

	for _, line1 := range newlines {
		key1, _ := splitIntoTwo(line1, ":")

		for j, line2 := range header {
			key2, _ := splitIntoTwo(line2, ":")

			if key1 == key2 {
				p.Content[j] = line1
				break
			}
		}
	}

	p.Stage = "edited"
	UpdateDiff(p)
	return true, p
}

func (f *Filesystem) Doom(p *Page, hard bool) (bool, *Page) {

	p.Stage = "doomed"
	return true, p
}

func ensureOriginalContent(p *Page) {
	if p == nil || p.Path == "" || p.Stage == "draft" {
		return
	}

	if hasOriginalContent(p.Og) {
		return
	}

	_, _, abspath := getPaths(p.Path)
	if p.Type == "deep" {
		abspath = filepath.Join(abspath, "index")
	}

	content := getContent(abspath)
	if len(content) == 0 {
		return
	}

	if p.Og == nil {
		p.Og = &Page{}
	}

	if p.Og.Name == "" {
		p.Og.Name = p.Name
	}
	if p.Og.Path == "" {
		p.Og.Path = p.Path
	}
	if p.Og.Type == "" {
		p.Og.Type = p.Type
	}
	if p.Og.Stage == "" {
		p.Og.Stage = "local"
	}

	p.Og.Content = append([]string(nil), content...)
}

func hasOriginalContent(p *Page) bool {
	return p != nil && len(p.Content) > 0
}
