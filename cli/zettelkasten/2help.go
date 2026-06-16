package zettelkasten

import (
	"env/filesystem"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func (z *Zettelkasten) rebuild() {
	selectedTag := z.selectedTagName()
	selectedOverlap := z.selectedOverlapName()

	if zettelDebugData {
		z.RandPage = zettelDebugRandPage()
		z.FamiPage = zettelDebugFamiPage()
	}

	z.rebuildTagPaths()
	z.rebuildFamilyTitles()
	z.rebuildTags()
	z.rebuildOverlaps()

	z.selectTag(selectedTag)
	z.selectOverlap(selectedOverlap)
	z.clampSelections()
}

func (z *Zettelkasten) rebuildTagPaths() {
	z.TagPaths = map[string][]string{}
	for _, page := range flattenPages(z.RandPage) {
		if page == nil || page.Type == "deep" {
			continue
		}
		for _, tag := range metadataTags(page) {
			z.TagPaths[tag] = appendUniqueString(z.TagPaths[tag], page.Path)
		}
	}
}

func (z *Zettelkasten) rebuildFamilyTitles() {
	z.FamilyTitles = map[string][]string{}
	if z.FamiPage == nil {
		return
	}

	for _, family := range z.FamiPage.Children {
		if family == nil {
			continue
		}
		z.FamilyTitles[family.Path] = familyDescendantTitles(family)
	}
}

func (z *Zettelkasten) rebuildTags() {
	z.Tags = []*ZettelTagItem{}

	filter := z.promptFilter()

	for tag, paths := range z.TagPaths {
		if filter != "" && !strings.Contains(strings.ToLower(tag), filter) {
			continue
		}

		z.Tags = append(z.Tags, &ZettelTagItem{
			Name:  tag,
			Paths: append([]string{}, paths...),
		})
	}

	sort.SliceStable(z.Tags, func(i, j int) bool {
		if len(z.Tags[i].Paths) != len(z.Tags[j].Paths) {
			return len(z.Tags[i].Paths) > len(z.Tags[j].Paths)
		}
		return strings.ToLower(z.Tags[i].Name) < strings.ToLower(z.Tags[j].Name)
	})
}

func (z *Zettelkasten) rebuildOverlaps() {
	z.Overlaps = []*ZettelOverlapItem{}

	selected := z.selectedTagName()
	if selected == "" {
		return
	}

	selectedPaths := z.TagPaths[selected]
	if len(selectedPaths) == 0 {
		return
	}

	for tag, paths := range z.TagPaths {
		if tag == selected {
			continue
		}

		shared := overlapCount(selectedPaths, paths)
		if shared == 0 {
			continue
		}

		z.Overlaps = append(z.Overlaps, &ZettelOverlapItem{
			Name:         tag,
			Paths:        append([]string{}, paths...),
			Shared:       shared,
			Percent:      int(float64(shared) / float64(len(selectedPaths)) * 100),
			Popularity:   len(paths),
			SelectedBase: len(selectedPaths),
		})
	}

	sort.SliceStable(z.Overlaps, func(i, j int) bool {
		if z.Overlaps[i].Percent != z.Overlaps[j].Percent {
			return z.Overlaps[i].Percent > z.Overlaps[j].Percent
		}
		if z.Overlaps[i].Shared != z.Overlaps[j].Shared {
			return z.Overlaps[i].Shared > z.Overlaps[j].Shared
		}
		if z.Overlaps[i].Popularity != z.Overlaps[j].Popularity {
			return z.Overlaps[i].Popularity > z.Overlaps[j].Popularity
		}

		return strings.ToLower(z.Overlaps[i].Name) < strings.ToLower(z.Overlaps[j].Name)
	})
}

func overlapCount(a []string, b []string) int {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}

	seen := map[string]bool{}
	for _, path := range a {
		seen[path] = true
	}

	count := 0
	for _, path := range b {
		if seen[path] {
			count++
		}
	}
	return count
}

func zettelPageTime(page *filesystem.Page) (time.Time, bool) {
	if page == nil || page.Metadata == nil {
		return time.Time{}, false
	}

	if t, ok := page.Metadata["time"].(time.Time); ok && !t.IsZero() {
		return t, true
	}

	if t, ok := page.Metadata["release-time"].(time.Time); ok && !t.IsZero() {
		return t, true
	}

	if raw, ok := page.Metadata["release-time"].(string); ok {
		layouts := []string{
			"02.01.2006 15:04",
			"02.01.2006",
			"2006-01-02 15:04",
			"2006-01-02",
			time.RFC3339,
		}

		for _, layout := range layouts {
			t, err := time.ParseInLocation(layout, strings.TrimSpace(raw), time.Local)
			if err == nil && !t.IsZero() {
				return t, true
			}
		}
	}

	return time.Time{}, false
}

func metadataTags(page *filesystem.Page) []string {
	if page == nil || page.Metadata == nil {
		return nil
	}

	return normalizeTags(page.Metadata["tags"])
}

func normalizeTags(raw any) []string {
	out := []string{}
	add := func(s string) {
		s = strings.TrimSpace(s)
		s = strings.TrimPrefix(s, "#")
		s = strings.ToLower(strings.TrimSpace(s))
		if s == "" {
			return
		}
		out = appendUniqueString(out, s)
	}

	switch v := raw.(type) {
	case []string:
		for _, s := range v {
			add(s)
		}
	case []any:
		for _, item := range v {
			add(fmt.Sprint(item))
		}
	case string:
		parts := strings.FieldsFunc(v, func(r rune) bool {
			return r == ',' || r == ' ' || r == '\t' || r == ';'
		})
		for _, part := range parts {
			add(part)
		}
	}

	return out
}

func familyDescendantTitles(root *filesystem.Page) []string {
	out := []string{}
	var walk func(*filesystem.Page)
	walk = func(page *filesystem.Page) {
		if page == nil {
			return
		}
		for _, child := range page.Children {
			if child == nil {
				continue
			}
			out = append(out, pageTitle(child))
			walk(child)
		}
	}
	walk(root)
	return out
}

func flattenPages(root *filesystem.Page) []*filesystem.Page {
	out := []*filesystem.Page{}
	var walk func(*filesystem.Page)
	walk = func(page *filesystem.Page) {
		if page == nil {
			return
		}
		out = append(out, page)
		for _, child := range page.Children {
			walk(child)
		}
	}
	walk(root)
	return out
}

func findZettelPage(page *filesystem.Page, path string) *filesystem.Page {
	if page == nil {
		return nil
	}
	if filepath.Clean(page.Path) == filepath.Clean(path) {
		return page
	}
	for _, child := range page.Children {
		found := findZettelPage(child, path)
		if found != nil {
			return found
		}
	}
	return nil
}

func pageTitle(page *filesystem.Page) string {
	if page == nil {
		return ""
	}
	if page.Metadata != nil {
		if name, ok := page.Metadata["name"].(string); ok && strings.TrimSpace(name) != "" {
			return strings.TrimSpace(name)
		}
	}
	if strings.TrimSpace(page.Name) != "" {
		return page.Name
	}
	return filepath.Base(page.Path)
}

func appendUniqueString(list []string, value string) []string {
	for _, item := range list {
		if item == value {
			return list
		}
	}
	return append(list, value)
}

func (z *Zettelkasten) selectedTagName() string {
	if z == nil || z.SelectedTag < 0 || z.SelectedTag >= len(z.Tags) || z.Tags[z.SelectedTag] == nil {
		return ""
	}
	return z.Tags[z.SelectedTag].Name
}

func (z *Zettelkasten) selectedOverlapName() string {
	item := z.selectedOverlap()
	if item == nil {
		return ""
	}
	return item.Name
}

func (z *Zettelkasten) selectedOverlap() *ZettelOverlapItem {
	if z == nil || z.SelectedOverlap < 0 || z.SelectedOverlap >= len(z.Overlaps) {
		return nil
	}
	return z.Overlaps[z.SelectedOverlap]
}

func (z *Zettelkasten) selectTag(name string) {
	if name == "" {
		z.clampSelections()
		return
	}
	for i, item := range z.Tags {
		if item != nil && item.Name == name {
			z.SelectedTag = i
			z.rebuildOverlaps()
			z.clampSelections()
			return
		}
	}
	z.clampSelections()
}

func (z *Zettelkasten) selectOverlap(name string) {
	if name == "" {
		z.clampSelections()
		return
	}
	for i, item := range z.Overlaps {
		if item != nil && item.Name == name {
			z.SelectedOverlap = i
			return
		}
	}
	z.clampSelections()
}

func (z *Zettelkasten) clampSelections() {
	if z.SelectedTag < 0 {
		z.SelectedTag = 0
	}
	if z.SelectedTag >= len(z.Tags) {
		z.SelectedTag = len(z.Tags) - 1
	}
	if z.SelectedTag < 0 {
		z.SelectedTag = 0
	}

	if z.SelectedOverlap < 0 {
		z.SelectedOverlap = 0
	}
	if z.SelectedOverlap >= len(z.Overlaps) {
		z.SelectedOverlap = len(z.Overlaps) - 1
	}
	if z.SelectedOverlap < 0 {
		z.SelectedOverlap = 0
	}
}

func cutLines(height, selected, total int) int {
	if height <= 0 || total <= height {
		return 0
	}
	start := selected - height/2
	if start < 0 {
		start = 0
	}
	if start+height > total {
		start = total - height
	}
	if start < 0 {
		start = 0
	}
	return start
}

func (z *Zettelkasten) promptFilter() string {
	if z == nil {
		return ""
	}

	filter := strings.TrimSpace(z.Prompt)
	filter = strings.TrimPrefix(filter, "#")
	filter = strings.ToLower(filter)

	return filter
}

func (z *Zettelkasten) familyMatchesPrompt(family *filesystem.Page, filter string) bool {
	if family == nil {
		return false
	}

	if filter == "" {
		return true
	}

	name := strings.ToLower(pageTitle(family))
	path := strings.ToLower(family.Path)

	if strings.Contains(name, filter) || strings.Contains(path, filter) {
		return true
	}

	for _, title := range z.FamilyTitles[family.Path] {
		if strings.Contains(strings.ToLower(title), filter) {
			return true
		}
	}

	return false
}
