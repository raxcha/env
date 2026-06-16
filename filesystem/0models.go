package filesystem

import "sync"

type ApiClient interface {
	Fetch(path string, depth int) (*Page, error)
	Push(p *Page) error
	Delete(path string) error
}

type Filesystem struct {
	mu          sync.Mutex
	Api         ApiClient
	Cache       *Page
	Page        chan *Page
	LoadRequest chan *LoadRequest
	SyncRequest chan *SyncRequest
	LogRequest  chan *LogRequest
}

type Page struct {
	Name string
	Path string
	Type string

	Options  map[string]any
	Content  []string
	Metadata map[string]any

	Children []*Page
	Sorting  string

	Og    *Page
	Stage string
	Diff  []string
}

type LoadRequest struct {
	Api              bool
	Mode             string
	Filter           string
	Path             string
	Opts, Cont, Meta bool
	Depth            int
	Sort             string
}

type SyncRequest struct {
	Api      bool
	Branch   *Page
	Filter   string
	Hard     bool
	PageOnly bool
}

type LogRequest struct {
	Page  *Page
	Index int
}

type log struct {
	typee     string
	text      []string
	timestamp string
	kind      string
	hidden    bool
	name      string
	id        string
	tags      []string
	refs      []string
	title     string
	path      string
	metadata  []string
	content   []string
}
