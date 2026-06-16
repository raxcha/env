package filesystem

import (
	"strings"
	"time"
)

func (f *Filesystem) Log(req *LogRequest) {
	f.LogRequest <- req
}

func (f *Filesystem) loggingRoudabout() {

	go func() {
		for req := range f.LogRequest {

			typee, text := getText(req)
			switch typee {
			case "none":
				f.Page <- errorPage
				continue
				
			case "block":
				f.mu.Lock()
				_, result := f.block(req, text)
				f.mu.Unlock()
				f.Page <- result
				continue

			case "snippet":
				f.mu.Lock()
				_, result := f.snippet(req, text)
				f.mu.Unlock()
				f.Page <- result
				continue
			}
		}
	}()
}

func (f *Filesystem) snippet(req *LogRequest, text []string) (ok bool, ptr *Page) {

	info := &log{}
	info.typee = "snippet"
	info.text = text

	info.timestamp = extractTimestamp(req.Page.Content[:req.Index])

	var message string
	ok, info.kind, info.id, message = extractSnippet(text[0])
	if !ok { return false, errorPage }

	info.refs, info.tags = extractTagsAndRefs([]string{message})
	info.name = message

	ptr = f.mountSnippet(info)
	f.merge(ptr)
	return true, f.point(ptr)
}

func (f *Filesystem) mountSnippet(info *log) *Page {

	path := "c.rand/.reminders"
	if info.kind == "~~" {
		path = "b.rec/.reminders"
	}

	page := findPointer(f.Cache, path)
	if page == nil {
		return errorPage
	}

	idTag := "#" + info.id
	for _, line := range page.Content {
		for _, field := range strings.Fields(line) {
			if field == idTag {
				return page
			}
		}
	}

	block := []string{
		info.kind + " " + info.name,
		idTag,
		"",
	}
	page.Content = append(block, page.Content...)
	page.Stage = "edited"

	return page
}

func extractSnippet(line string) (ok bool, kind, id, message string) {
	parts := strings.SplitN(line, "|", 2)
	if len(parts) != 2 {
		return false, "", "", ""
	}

	left := strings.Fields(parts[0])
	if len(left) < 2 {
		return false, "", "", ""
	}

	ok, kind = extractKind(left[0])
	if !ok || len(kind) != 2 {
		return false, "", "", ""
	}

	id = left[1]
	if len(id) < 8 {
		return false, "", "", ""
	}

	message = strings.TrimSpace(parts[1])
	if message == "" {
		return false, "", "", ""
	}

	return true, kind, id, message
}



func (f *Filesystem) block(req *LogRequest, text []string) (ok bool, ptr *Page) {

	info := &log{}
	info.typee = "block"
	info.text = text
	
	info.timestamp = extractTimestamp(req.Page.Content[:req.Index])

	ok, info.kind = extractKind(text[0])
	if !ok { return false, errorPage }

	info.hidden = extractHidden(text[0])

	ok, info.name = extractName(text[0])
	if !ok { return false, errorPage }

	ok, info.id = extractId(text[1])
	if !ok { return false, errorPage }

	info.refs, info.tags = extractTagsAndRefs(text)
	if info.kind == "`" && len(info.tags) == 0 {
		return false, errorPage 
	} else if len(info.refs) == 0 { return false, errorPage }

	ptr = f.mountBlock(info)
	f.merge(ptr)
	return true, f.point(ptr)

}

func (f *Filesystem) mountBlock(info *log) *Page {

	quotes := ""
	ref := ""
	kind := ""
	path := ""
	switch info.kind {
	case ">":
		quotes = "{}"
		ref = info.refs[0]
		kind = "plan"
		path = "b.rec/"
	case "=":
		quotes = "[]"
		ref = info.refs[0]
		kind = "session"
		path = "b.rec/"
	case "~":
		quotes = "()"
		ref = info.refs[0]
		kind = "task"
		path = "b.rec/"
	case "`":
		quotes = "<>"
		ref = info.tags[0]
		kind = "idea"
		path = "c.rand/"
	}
	
	hidden := "no"
	if info.hidden { hidden = "yes" }

	info.title = info.kind + info.name + string(quotes[0]) + ref + string(quotes[1])
	info.metadata = []string{
		"id: " + info.id,
		"type: " + kind,
		"name: " + info.name,
		"tags: " + strings.Join(info.tags, ", "),
		"references: " + strings.Join(info.refs, ", "),
		"hidden: " + hidden,
		"release-time: " + info.timestamp,
		"author: nausea",
		"status: ",
		"last-edited-time: ",
		"---",
	}

	info.path = path + info.title

	if info.hidden {
		info.content = info.text[2:]
	} else {
		info.content = []string{""}
	}
	
	page := newPage()
	page.Type = "shallow"
	page.Content = append(info.metadata, info.content...)
	page.Name = info.title
	page.Path = info.path
	page.Sorting = "basic"
	page.Stage = "edited"
	page.Og = snapPage(page)
	return page
}

func extractTagsAndRefs(lines []string) (mentions []string, tags []string) {
    for _, line := range lines {
        fields := strings.Fields(line)
        for _, f := range fields {
            if strings.HasPrefix(f, "@") && len(f) > 1 {
                mentions = append(mentions, f[1:]) // sem o @
            } else if strings.HasPrefix(f, "#") && len(f) > 1 {
                tags = append(tags, f[1:]) // sem o #
            }
        }
    }
    return
}

func extractId(text string) (bool, string) {
    fields := strings.Fields(text)
    
    for i, f := range fields {
        if strings.ToLower(f) == "id:" && i+1 < len(fields) {
            id := fields[i+1]
            if len(id) >= 12 {
                return true, id
            }
        }
    }
    
    return false, ""
}

func extractName(text string) (bool, string) {

    fields := strings.Fields(text)
    if len(fields) < 2 { return false, "" }
    start := 1
    if fields[start] == "!" { start++ }
    var result []string
    for _, f := range fields[start:] {
        if strings.HasPrefix(f, "@") || strings.HasPrefix(f, "#") { break }
        result = append(result, f)
    }

    return true, strings.Join(result, "-")
}
func extractHidden(text string) bool {

    fields := strings.Fields(text)
    if len(fields) < 2 { return false }
    symbol := fields[0]
    bang := fields[1]
    if len(symbol) < 1 || len(symbol) > 2 { return false }
    return bang == "!"
}

func extractKind(text string) (bool, string) {

	if len(text) < 3 { return false, "" }
	if strings.ContainsRune("-`~>=", rune(text[0])) {
		if strings.ContainsRune("-~", rune(text[0])) && strings.ContainsRune("-~", rune(text[1])) {
			return true, text[:2]
		}
		return true, string(text[0])
	}
	return false, ""
}

func extractTimestamp(lines []string) string {
    var date, t string

    for i := len(lines) - 1; i >= 0; i-- {
        fields := strings.Fields(lines[i])

        var foundDate, foundTime string
        for _, f := range fields {
            if foundTime == "" {
                if _, err := time.Parse("15:04", f); err == nil {
                    foundTime = f
                }
            }
            if foundDate == "" {
                if _, err := time.Parse("02.01.2006", f); err == nil {
                    foundDate = f
                }
            }
        }

        if foundTime != "" {
            t = foundTime
        }
        if foundDate != "" {
            date = foundDate
            if t == "" {
                t = "00:00"
            }
            return date + " " + t
        }
    }

    return time.Now().Format("02.01.2006 15:04")
}

func getText(req *LogRequest) (string, []string) {

	lines := req.Page.Content
	start := -1
	for i := req.Index ; i >= 0 ; i-- {
		if isSeparatorLine(lines[i]) { start = i ; break }
	}
	if start == -1 { return "none", []string{} }
	for i := req.Index ; i < len(lines) ; i++ {
		if isSeparatorLine(lines[i]) || i == len(lines)-1 {
			if i - start == 1 { return "snippet", lines[start:i]}
			if i - start < 3 { return "none", []string{} }
			return "block", lines[start:i]
		} 
	}
	return "none", []string{}
}