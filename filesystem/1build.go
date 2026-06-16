package filesystem

var cache = &Page{
	Name:     "prsnl.spc",
	Path:     ".",
	Type:     "cache",
	Options:  map[string]any{},
	Content:  []string{},
	Metadata: map[string]any{},
	Children: []*Page{},
	Sorting:  "basic",
	Og:       &Page{},
	Stage:    "ghost",
}

var errorPage = &Page{
	Name:     "error",
	Path:     "error",
	Type:     "error",
	Options:  map[string]any{"error": "error"},
	Content:  []string{"error"},
	Metadata: map[string]any{"error": "error"},
	Children: []*Page{},
	Sorting:  "error",
	Og:       &Page{},
	Stage:    "error",
}

func CreateFilesystem() *Filesystem {

	f := Filesystem{
		Cache:       cache,
		Page:        make(chan *Page, 10),
		LoadRequest: make(chan *LoadRequest, 10),
		SyncRequest: make(chan *SyncRequest, 10),
		LogRequest:  make(chan *LogRequest, 10),
	}
	f.loadingRoudabout()
	f.syncingRoudabout()
	f.loggingRoudabout()

	return &f
}

func newPage() *Page {

	return &Page{
		Name:     "",
		Path:     "",
		Type:     "",
		Options:  map[string]any{},
		Content:  []string{},
		Metadata: map[string]any{},
		Children: []*Page{},
		Sorting:  "basic",
		Og:       &Page{},
		Stage:    "",
		Diff:     []string{},
	}
}

func NewLoadRequest() *LoadRequest {

	return &LoadRequest{
		Api:    false,
		Mode:   "",
		Filter: "",
		Path:   "",
		Opts:   true,
		Cont:   true,
		Meta:   true,
		Depth:  -1,
		Sort:   "",
	}
}

func NewSyncRequest() *SyncRequest {

	return &SyncRequest{
		Api:      false,
		Branch:   &Page{},
		Filter:   "",
		Hard:     false,
		PageOnly: false,
	}
}
