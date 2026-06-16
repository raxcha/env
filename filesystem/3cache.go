package filesystem

import (
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func snapPage(p *Page) *Page {
	content := make([]string, len(p.Content))
	copy(content, p.Content)
	return &Page{
		Name:     p.Name,
		Path:     p.Path,
		Type:     p.Type,
		Options:  p.Options,
		Content:  content,
		Metadata: p.Metadata,
		Stage:    p.Stage,
		Sorting:  p.Sorting,
		Children: []*Page{},
		Og:       &Page{},
	}
}

func findPointer(page *Page, path string) *Page {

	if page == nil {
		return nil
	}
	if page.Path == path {
		return page
	}
	for _, child := range page.Children {
		res := findPointer(child, path)
		if res != nil {
			return res
		}
	}
	return nil
}

func sortPages(list []*Page, mode string) []*Page {

	sort.Slice(list, func(i, j int) bool {

		a := list[i]
		b := list[j]

		switch mode {

		case "depth":
			return calculatePathDepth(a.Path) > calculatePathDepth(b.Path)

		case "depth2":
			return calculatePathDepth(a.Path) < calculatePathDepth(b.Path)

		case "basic":
			if a.Type != b.Type {
				return a.Type == "deep"
			}
			return strings.ToLower(a.Name) < strings.ToLower(b.Name)

		case "priority":
			aPriority, _ := a.Metadata["priority"].(int)
			bPriority, _ := b.Metadata["priority"].(int)
			if aPriority != bPriority {
				return aPriority > bPriority
			}
			return strings.ToLower(list[i].Name) < strings.ToLower(list[j].Name)

		case "time":
			aTime, _ := a.Metadata["time"].(time.Time)
			bTime, _ := b.Metadata["time"].(time.Time)
			if !aTime.Equal(bTime) {
				return aTime.After(bTime)
			}
			return strings.ToLower(list[i].Name) < strings.ToLower(list[j].Name)

		default:
			return false

		}
	})

	return list

}

func sortPage(page *Page, mode string) *Page {

	if page == nil || len(page.Children) == 0 {
		return page
	}

	sortPages(page.Children, mode)
	for _, child := range page.Children {
		sortPage(child, mode)
	}

	return page
}

func nestPage(list []*Page) *Page {

	list = sortPages(list, "depth2")
	mapp := mapPages(list)

	if mapp["."] == nil {
		root := newPage()
		root.Name = "."
		root.Path = "."
		root.Type = "deep"
		mapp["."] = root
	}

	for _, p := range list {
		if p == nil {
			continue
		}

		p.Children = nil
	}

	for _, p := range list {

		if p == nil || p.Path == "." {
			continue
		}

		parent := mapp[filepath.Dir(p.Path)]

		if parent == nil {
			parent = mapp["."]
		}

		parent.Children = append(parent.Children, p)
	}

	return mapp["."]
}

func mapPages(list []*Page) map[string]*Page {

	mapp := map[string]*Page{}

	if mapp["."] == nil {
		root := newPage()
		root.Name = "."
		root.Path = "."
		root.Type = "deep"
		mapp["."] = root
	}

	for _, p := range list {
		mapp[p.Path] = p
	}
	return mapp
}

func (f *Filesystem) merge(page *Page) {

	if page == nil {
		return
	}

	node := findPointer(f.Cache, page.Path)
	if node == nil {
		node = insertPointer(f.Cache, page)
	} else {
		node = replacePage(node, page)
	}

	for _, child := range page.Children {
		f.merge(child)
	}
}

func insertPointer(root *Page, page *Page) *Page {

	parts := strings.Split(page.Path, "/")

	current := root

	for _, part := range parts {

		if part == "." || part == "" {
			continue
		}

		found := false
		for _, child := range current.Children {
			if child.Name == part {
				current = child
				found = true
				break
			}
		}

		if !found {
			newchild := newPage()
			newchild.Name = part
			newchild.Path = filepath.Join(current.Path, part)

			current.Children = append(current.Children, newchild)

			current = newchild
		}
	}

	current = replacePage(current, page)

	if len(page.Children) > 0 {
		current.Children = page.Children
	}

	return current

}

func (f *Filesystem) point(page *Page) *Page {

	return findPointer(f.Cache, page.Path)
}

func (f *Filesystem) Find(path string) *Page {

	f.mu.Lock()
	defer f.mu.Unlock()
	return findPointer(f.Cache, path)
}

func replacePage(old *Page, new *Page) *Page {

	if canReplaceCachedPage(old, new) {
		old.Name = new.Name
		old.Path = new.Path
		old.Type = new.Type
		old.Options = new.Options
		if len(new.Content) > 0 || len(old.Content) == 0 {
			old.Content = new.Content
		}
		old.Metadata = new.Metadata
		old.Stage = new.Stage
		old.Og = mergeOriginalPage(old.Og, new.Og)
	}

	return old
}

func mergeOriginalPage(old, new *Page) *Page {
	if new == nil {
		return old
	}

	if old == nil {
		return new
	}

	if len(new.Content) == 0 && len(old.Content) > 0 {
		new.Content = old.Content
	}

	return new
}

func canReplaceCachedPage(old *Page, new *Page) bool {

	if old == nil || new == nil {
		return false
	}

	// "" is an empty placeholder (lowest quality) — any stage can replace it
	if old.Stage == "" {
		return true
	}

	if new.Stage == "ghost" && old.Stage != "ghost" {
		return false
	}
	if new.Stage == "local" && old.Stage != "ghost" && old.Stage != "local" {
		return false
	}
	if new.Stage == "api" && old.Stage != "ghost" && old.Stage != "local" && old.Stage != "api" {
		return false
	}
	return true
}

func removePage(root *Page, path string) *Page {

	for i, child := range root.Children {
		if child.Path == path {
			root.Children = append(root.Children[:i], root.Children[i+1:]...)
			return root
		}
		removePage(child, path)
	}
	return root
}

func findParent(page *Page, path string) *Page {

	parentpath := filepath.Dir(path)
	return findPointer(page, parentpath)
}

func UpdateDiff(p *Page) {

	ensureOriginalContent(p)

	og := p.Og
	if og == nil {
		p.Diff = []string{}
		return
	}

	var out []string

	if og.Name != p.Name {
		out = append(out, "name:  "+og.Name+" → "+p.Name)
	}
	if og.Path != p.Path {
		out = append(out, "path:  "+og.Path+" → "+p.Path)
	}
	if og.Type != p.Type {
		out = append(out, "type:  "+og.Type+" → "+p.Type)
	}
	if og.Stage != p.Stage {
		out = append(out, "stage: "+og.Stage+" → "+p.Stage)
	}

	if hasOriginalContent(og) || len(p.Content) == 0 {
		cd := diffLines(og.Content, p.Content)
		if len(cd) > 0 {
			out = append(out, "---")
			out = append(out, cd...)
		}
	}

	p.Diff = out
}

func DiffLines(a, b []string) []string {
	return diffLines(a, b)
}

func FullDiffLines(a, b []string) []string {
	return diffLinesInternal(a, b, true)
}

func diffLines(a, b []string) []string {
	return diffLinesInternal(a, b, false)
}

func diffLinesInternal(a, b []string, context bool) []string {
	m, n := len(a), len(b)

	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] >= dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}

	var out []string
	i, j := m, n
	for i > 0 || j > 0 {
		switch {
		case i > 0 && j > 0 && a[i-1] == b[j-1]:
			if context {
				out = append([]string{"  " + a[i-1]}, out...)
			}
			i--
			j--
		case j > 0 && (i == 0 || dp[i][j-1] >= dp[i-1][j]):
			out = append([]string{"+ " + b[j-1]}, out...)
			j--
		default:
			out = append([]string{"- " + a[i-1]}, out...)
			i--
		}
	}
	return out
}

func RepathPage(page *Page, newpath string) {
	oldPath := page.Path
	page.Path = newpath

	for _, child := range page.Children {
		childRelative := strings.TrimPrefix(child.Path, oldPath)
		child.Path = newpath + childRelative
		RepathPage(child, child.Path)
	}
}

func (f *Filesystem) PromoteChildren(target *Page) {
	parent := findParent(f.Cache, target.Path)
	if parent == nil {
		return
	}

	for _, child := range target.Children {
		child.Path = filepath.Join(filepath.Dir(target.Path), child.Name)
		RepathPage(child, child.Path)
		parent.Children = append(parent.Children, child)
	}

	target.Children = nil
}
