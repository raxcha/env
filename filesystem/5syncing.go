package filesystem

import (
	"os"
	"path/filepath"
	"strings"
)

func (f *Filesystem) Sync(req *SyncRequest) {
	f.SyncRequest <- req
}

func (f *Filesystem) syncingRoudabout() {

	go func() {
		for req := range f.SyncRequest {

			if req.Api {
				if f.Api == nil {
					f.Page <- errorPage
					continue
				}
				f.mu.Lock()
				_, newptr := f.syncApi(req)
				f.mu.Unlock()
				f.Page <- newptr
				continue
			}

			f.mu.Lock()
			_, newptr := f.sync(req)
			f.mu.Unlock()
			f.Page <- newptr
		}
	}()
}

func (f *Filesystem) sync(req *SyncRequest) (ok bool, newptr *Page) {

	doomed := []string{}
	if req.PageOnly {
		if syncPage(req.Branch, parseFilter(req.Filter)) {
			doomed = append(doomed, req.Branch.Path)
		}
	} else {
		doomed = syncBranch(req.Branch, parseFilter(req.Filter))
	}

	for _, path := range doomed {
		f.Cache = removePage(f.Cache, path)
	}

	newptr = findPointer(f.Cache, req.Branch.Path)
	if newptr == nil {
		return false, errorPage
	}
	return true, newptr
}

func (f *Filesystem) syncApi(req *SyncRequest) (ok bool, newptr *Page) {

	doomed := []string{}
	if req.PageOnly {
		if syncApiPage(req.Branch, f.Api, parseFilter(req.Filter)) {
			doomed = append(doomed, req.Branch.Path)
		}
	} else {
		doomed = syncApiBranch(req.Branch, f.Api, parseFilter(req.Filter))
	}

	for _, path := range doomed {
		f.Cache = removePage(f.Cache, path)
	}

	newptr = findPointer(f.Cache, req.Branch.Path)
	if newptr == nil {
		return false, errorPage
	}
	return true, newptr
}

func syncApiBranch(page *Page, client ApiClient, filter func(*Page) bool) (doomed []string) {

	for _, child := range page.Children {
		doomed = append(doomed, syncApiBranch(child, client, filter)...)
	}

	if filter != nil && !filter(page) {
		return
	}

	switch page.Stage {
	case "doomed":
		if err := client.Delete(page.Path); err == nil {
			doomed = append(doomed, page.Path)
		}
	case "draft", "edited", "local":
		if err := client.Push(page); err == nil {
			page.Stage = "api"
			page.Og = snapPage(page)
			page.Diff = []string{}
		}
	}

	return
}

func syncApiPage(page *Page, client ApiClient, filter func(*Page) bool) bool {
	if page == nil {
		return false
	}

	if filter != nil && !filter(page) {
		return false
	}

	switch page.Stage {
	case "doomed":
		return client.Delete(page.Path) == nil
	case "draft", "edited", "local":
		if err := client.Push(page); err == nil {
			page.Stage = "api"
			page.Og = snapPage(page)
			page.Diff = []string{}
		}
	}

	return false
}

func syncBranch(page *Page, filter func(*Page) bool) (doomed []string) {

	for _, child := range page.Children {
		doomed = append(doomed, syncBranch(child, filter)...)
	}

	if filter != nil && !filter(page) {
		return
	}

	_, _, abspath := getPaths(page.Path)

	switch page.Stage {

	case "doomed":
		if page.Type == "deep" {
			os.RemoveAll(abspath)
		} else {
			os.Remove(abspath)
		}
		doomed = append(doomed, page.Path)

	case "draft", "edited", "api":
		if page.Type == "deep" {
			os.MkdirAll(abspath, 0755)
			os.WriteFile(filepath.Join(abspath, "index"), []byte(strings.Join(page.Content, "\n")), 0644)
		} else {
			os.WriteFile(abspath, []byte(strings.Join(page.Content, "\n")), 0644)
		}
		page.Stage = "local"
		page.Og = snapPage(page)
		page.Diff = []string{}
	}

	return
}

func syncPage(page *Page, filter func(*Page) bool) bool {
	if page == nil {
		return false
	}

	if filter != nil && !filter(page) {
		return false
	}

	_, _, abspath := getPaths(page.Path)

	switch page.Stage {
	case "doomed":
		if page.Type == "deep" {
			os.RemoveAll(abspath)
		} else {
			os.Remove(abspath)
		}
		return true

	case "draft", "edited", "api":
		if page.Type == "deep" {
			os.MkdirAll(abspath, 0755)
			os.WriteFile(filepath.Join(abspath, "index"), []byte(strings.Join(page.Content, "\n")), 0644)
		} else {
			os.WriteFile(abspath, []byte(strings.Join(page.Content, "\n")), 0644)
		}
		page.Stage = "local"
		page.Og = snapPage(page)
		page.Diff = []string{}
	}

	return false
}
