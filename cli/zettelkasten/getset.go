package zettelkasten

import (
	"env/engine"
	"env/filesystem"
	"path/filepath"
	"strings"
)

func (z *Zettelkasten) GetName() string               { return z.Name }
func (z *Zettelkasten) GetPath() string               { return z.Path }
func (z *Zettelkasten) GetKind() string               { return "zettelkasten" }
func (z *Zettelkasten) GetFloating() bool             { return z.Float }
func (z *Zettelkasten) GetBounds() *engine.Boundaries { return &z.Bounds }

func (z *Zettelkasten) SetPage(page *filesystem.Page) {
	if z == nil || page == nil {
		return
	}

	selectedTag := z.selectedTagName()
	selectedOverlap := z.selectedOverlapName()

	clean := filepath.ToSlash(filepath.Clean(page.Path))

	if clean == "c.rand" || strings.HasPrefix(clean, "c.rand/") {
		if clean == "c.rand" {
			z.RandPage = page
		} else if z.RandPage != nil {
			mergeZettelPage(z.RandPage, page)
		}

		z.rebuild()
		z.selectTag(selectedTag)
		z.selectOverlap(selectedOverlap)
		return
	}

	if clean == "d.fami" || strings.HasPrefix(clean, "d.fami/") {
		if clean == "d.fami" {
			z.FamiPage = page
		} else if z.FamiPage != nil {
			mergeZettelPage(z.FamiPage, page)
		}

		z.rebuild()
		z.selectTag(selectedTag)
		z.selectOverlap(selectedOverlap)
	}
}

func (z *Zettelkasten) AcceptsPage(page *filesystem.Page) bool {
	if z == nil || page == nil {
		return false
	}

	clean := filepath.ToSlash(filepath.Clean(page.Path))
	return clean == "c.rand" ||
		clean == "d.fami" ||
		strings.HasPrefix(clean, "c.rand/") ||
		strings.HasPrefix(clean, "d.fami/")
}

func mergeZettelPage(root *filesystem.Page, page *filesystem.Page) {
	if root == nil || page == nil {
		return
	}

	for i, child := range root.Children {
		if child == nil {
			continue
		}

		if filepath.Clean(child.Path) == filepath.Clean(page.Path) {
			root.Children[i] = page
			return
		}

		mergeZettelPage(child, page)
	}
}

func (z *Zettelkasten) GetSpec() string {
	return z.Path
}

func (z *Zettelkasten) GetTitle() string                     { return z.Name }
func (z *Zettelkasten) IsSidebarOn() bool                    { return false }
func (z *Zettelkasten) GetSidebarBounds() *engine.Boundaries { return nil }
