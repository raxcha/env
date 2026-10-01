package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"env/filesystem"
	"env/types"
)

func TestFilesystemProtocol(t *testing.T) {
	root := t.TempDir()
	f, err := filesystem.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(New(f, "user", "pass"))
	defer server.Close()
	client := filesystem.CreateFilesystem(true)
	client.Url = server.URL
	auth, err := http.Post(server.URL+"/auth", "application/json", bytes.NewBufferString(`{"username":"user","password":"pass"}`))
	if err != nil {
		t.Fatal(err)
	}
	var credentials struct{ Token string }
	if err = json.NewDecoder(auth.Body).Decode(&credentials); err != nil {
		t.Fatal(err)
	}
	auth.Body.Close()
	client.Token = credentials.Token
	page := &types.Page{Path: "notes/test", Type: "shallow", Stage: "draft", Content: []string{"hello", "world"}}
	ok, patch := client.SyncApi(&types.Syncing{Branch: page})
	if !ok {
		t.Fatal("SyncApi failed")
	}
	client.ApplyPatch(patch)
	if page.Stage != "api" {
		t.Fatalf("sync stage: %s", page.Stage)
	}
	data, err := os.ReadFile(filepath.Join(root, "notes/test"))
	if err != nil || string(data) != "hello\nworld" {
		t.Fatalf("disk: %q, %v", data, err)
	}
	req := client.NewLoading()
	req.Path = page.Path
	ok, loaded := client.LoadApi(req)
	if !ok || len(loaded.Content) != 2 || loaded.Og == nil {
		t.Fatalf("load: %#v", loaded)
	}
	page.Og = &types.Page{Path: page.Path}
	page.Path, page.Stage = "notes/moved", "move"
	_, patch = client.SyncApi(&types.Syncing{Branch: page})
	client.ApplyPatch(patch)
	if _, err = os.Stat(filepath.Join(root, "notes/test")); !os.IsNotExist(err) {
		t.Fatal("original remains after move")
	}
	page.Stage = "doom"
	_, patch = client.SyncApi(&types.Syncing{Branch: page})
	client.ApplyPatch(patch)
	if _, err = os.Stat(filepath.Join(root, "notes/moved")); !os.IsNotExist(err) {
		t.Fatal("file remains after delete")
	}
	for _, tc := range []struct {
		path, body, token string
		status            int
	}{
		{"/load", `{}`, "", 401},
		{"/auth", `{"username":"bad"}`, "", 401},
		{"/sync", `null`, client.Token, 400},
		{"/sync", `{"Path":"../escape","Stage":"draft"}`, client.Token, 400},
		{"/sync", `{"Path":".","Stage":"doom","Type":"deep"}`, client.Token, 400},
		{"/sync", `{`, client.Token, 400},
		{"/load", `{} {}`, client.Token, 400},
		{"/load", `{"Path":"missing"}`, client.Token, 404},
	} {
		request := httptest.NewRequest("POST", tc.path, bytes.NewBufferString(tc.body))
		request.Header.Set("Authorization", "Bearer "+tc.token)
		// Use the running handler, retaining its authentication token.
		response, err := func() (*http.Response, error) {
			request.URL.Scheme = "http"
			request.URL.Host = server.Listener.Addr().String()
			request.RequestURI = ""
			return http.DefaultClient.Do(request)
		}()
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != tc.status {
			t.Errorf("%s %s: got %d, want %d", tc.path, tc.body, response.StatusCode, tc.status)
		}
	}
}

func TestSessionLimit(t *testing.T) {
	f, err := filesystem.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	handler := New(f, "user", "pass")
	call := func(path, body, token string) *httptest.ResponseRecorder {
		t.Helper()
		request := httptest.NewRequest("POST", path, bytes.NewBufferString(body))
		request.Header.Set("Authorization", "Bearer "+token)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		return response
	}
	login := func() string {
		t.Helper()
		response := call("/auth", `{"username":"user","password":"pass"}`, "")
		var result struct{ Token string }
		if response.Code != 200 {
			t.Fatalf("login: %d", response.Code)
		}
		if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if result.Token == "" {
			t.Fatal("empty token")
		}
		return result.Token
	}
	check := func(token string, status int) {
		t.Helper()
		for _, path := range []string{"/load", "/sync"} {
			body := `{"Path":"."}`
			if path == "/sync" {
				body = `{"Path":"unused","Stage":"ghost"}`
			}
			want := status
			if path == "/sync" && status == 200 {
				want = 204
			}
			if got := call(path, body, token).Code; got != want {
				t.Errorf("%s: got %d, want %d", path, got, want)
			}
		}
	}
	check("", 401)
	check("unknown", 401)
	tokens := make([]string, 12)
	seen := map[string]bool{}
	for i := range tokens {
		tokens[i] = login()
		if seen[tokens[i]] {
			t.Fatal("repeated token")
		}
		seen[tokens[i]] = true
	}
	for _, token := range tokens {
		check(token, 200)
	}
	if got := call("/auth", `{"username":"user","password":"wrong"}`, "").Code; got != 401 {
		t.Fatalf("invalid login: %d", got)
	}
	for _, token := range tokens {
		check(token, 200)
	}
	check(tokens[0], 200) // Recent use does not change creation order.
	newest := login()
	check(tokens[0], 401)
	for _, token := range tokens[1:] {
		check(token, 200)
	}
	check(newest, 200)
	next := login()
	check(tokens[1], 401)
	for _, token := range tokens[2:] {
		check(token, 200)
	}
	check(newest, 200)
	check(next, 200)
	handler = New(f, "user", "pass")
	check(next, 401)
}

func TestShhh(t *testing.T) {
	root := t.TempDir()
	f, err := filesystem.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "notes/deep"), 0755); err != nil {
		t.Fatal(err)
	}
	expected := map[string]string{"notes/deep/test": "olá\nworld", "notes/index": "folder content", ".hidden": "hidden content"}
	for path, content := range expected {
		if err := os.WriteFile(filepath.Join(root, path), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, ".options"), []byte("password: secret"), 0600); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "secret")
	if err := os.WriteFile(outside, []byte("external secret"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}
	handler := New(f, "user", "pass")
	for _, tc := range []struct {
		name, user, pass string
		auth             bool
		status           int
	}{
		{"missing", "", "", false, 401},
		{"wrong password", "user", "bad", true, 401},
		{"wrong user", "bad", "pass", true, 401},
		{"valid", "user", "pass", true, 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/shhh", nil)
			if tc.auth {
				request.SetBasicAuth(tc.user, tc.pass)
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != tc.status {
				t.Fatalf("status: %d", response.Code)
			}
			if response.Header().Get("Cache-Control") != "no-store" {
				t.Fatal("export may be cached")
			}
			if tc.status == 401 {
				if response.Header().Get("WWW-Authenticate") == "" {
					t.Fatal("missing Basic challenge")
				}
				if bytes.Contains(response.Body.Bytes(), []byte("content")) {
					t.Fatal("unauthorized data")
				}
				return
			}
			var result struct {
				Files map[string]string `json:"files"`
			}
			if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if len(result.Files) != len(expected) {
				t.Fatalf("unexpected files: %v", result.Files)
			}
			for path, want := range expected {
				if result.Files[path] != want {
					t.Errorf("%s: %q", path, result.Files[path])
				}
			}
		})
	}
	for _, credentials := range [][2]string{{"", ""}, {"user", ""}, {"", "pass"}} {
		request := httptest.NewRequest("GET", "/shhh", nil)
		request.SetBasicAuth(credentials[0], credentials[1])
		response := httptest.NewRecorder()
		New(f, credentials[0], credentials[1]).ServeHTTP(response, request)
		if response.Code != 401 {
			t.Fatal("unconfigured credentials permit export")
		}
	}
	empty, err := filesystem.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest("GET", "/shhh", nil)
	request.SetBasicAuth("user", "pass")
	response := httptest.NewRecorder()
	New(empty, "user", "pass").ServeHTTP(response, request)
	if response.Code != 200 || response.Body.String() != "{\"files\":{}}\n" {
		t.Fatalf("empty export: %s", response.Body.String())
	}
}
