package filesystem

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"env/types"
)

// Open opens a server filesystem without starting client authentication or timers.
func Open(root string) (*Filesystem, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if err = os.MkdirAll(root, 0755); err != nil {
		return nil, err
	}
	f := &Filesystem{api: true, root: root, Standards: map[string][]string{}, Options: map[string]any{}, Cache: map[string]*types.Page{}, Patch: make(chan *types.Patch)}
	f.reload()
	return f, nil
}

// SavePage applies one remote page immediately. Children are sent separately by SyncApi.
// Unlike deferred UI patches, failures must reach the HTTP client before acknowledging sync.
func (f *Filesystem) SavePage(page *types.Page) error {
	if page == nil || page.Path == "" {
		return errors.New("page path is required")
	}
	root, err := os.OpenRoot(f.whereIsRoot())
	if err != nil {
		return err
	}
	defer root.Close()
	path := filepath.Clean(page.Path)
	if !filepath.IsLocal(path) {
		return errors.New("path must be relative to the filesystem root")
	}
	old := path
	if page.Og != nil {
		old = filepath.Clean(page.Og.Path)
	}
	if !filepath.IsLocal(old) {
		return errors.New("original path must be relative to the filesystem root")
	}
	switch page.Stage {
	case "ghost":
		return nil
	case "doom":
		if old == "." {
			return errors.New("cannot delete filesystem root")
		}
		if page.Type == "deep" {
			return root.RemoveAll(old)
		}
		return root.Remove(old)
	case "move":
		if path == "." || old == "." {
			return errors.New("cannot move filesystem root")
		}
		if old != path {
			if err = root.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return err
			}
			if err = root.Rename(old, path); err != nil {
				return err
			}
		}
	case "draft", "edit", "local", "*api", "api":
	default:
		return errors.New("unsupported page stage")
	}
	if page.Type == "deep" {
		path = filepath.Join(path, "index")
	}
	if err = root.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return root.WriteFile(path, []byte(strings.Join(page.Content, "\n")), 0644)
}

// ExportContent reads stored content without following symlinks or exporting client credentials.
func (f *Filesystem) ExportContent() (map[string]string, error) {
	root, err := os.OpenRoot(f.whereIsRoot())
	if err != nil {
		return nil, err
	}
	defer root.Close()
	files := make(map[string]string)
	err = fs.WalkDir(root.FS(), ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Name() == ".options" {
			if entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		data, err := root.ReadFile(path)
		if err != nil {
			return err
		}
		files[path] = string(data)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}
