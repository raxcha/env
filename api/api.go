// Package api exposes the existing filesystem client protocol over HTTP.
package api

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"env/filesystem"
	"env/types"
)

const maxSessions = 12

func New(f *filesystem.Filesystem, username, password string) http.Handler {
	var tokens []string
	var sessionsMu sync.Mutex
	var mu sync.Mutex
	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth", func(w http.ResponseWriter, r *http.Request) {
		var credentials struct{ Username, Password string }
		if !decode(w, r, &credentials) {
			return
		}
		if !equal(credentials.Username, username) || !equal(credentials.Password, password) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		token := rand.Text()
		sessionsMu.Lock()
		if len(tokens) == maxSessions {
			copy(tokens, tokens[1:])
			tokens = tokens[:maxSessions-1]
		}
		tokens = append(tokens, token)
		sessionsMu.Unlock()
		respond(w, map[string]string{"token": token})
	})
	protect := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			sessionsMu.Lock()
			valid := false
			for _, token := range tokens {
				if equal(r.Header.Get("Authorization"), "Bearer "+token) {
					valid = true
				}
			}
			sessionsMu.Unlock()
			if !valid {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			mu.Lock()
			defer mu.Unlock()
			next(w, r)
		}
	}
	mux.HandleFunc("POST /load", protect(func(w http.ResponseWriter, r *http.Request) {
		req := f.NewLoading()
		if !decode(w, r, req) {
			return
		}
		ok, page := f.LoadLocal(req)
		if !ok {
			http.Error(w, "load failed", http.StatusNotFound)
			return
		}
		if page != nil {
			page.Token = req.Token
		}
		respond(w, page)
	}))
	mux.HandleFunc("POST /sync", protect(func(w http.ResponseWriter, r *http.Request) {
		var page *types.Page
		if !decode(w, r, &page) {
			return
		}
		if page == nil || page.Path == "" {
			http.Error(w, "page path is required", http.StatusBadRequest)
			return
		}
		if err := f.SavePage(page); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	mux.HandleFunc("GET /shhh", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		user, pass, ok := r.BasicAuth()
		// Unlike local token authentication, exports require configured credentials.
		if username == "" || password == "" || !ok || !equal(user, username) || !equal(pass, password) {
			w.Header().Set("WWW-Authenticate", `Basic realm="prsnlspc", charset="UTF-8"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		mu.Lock()
		defer mu.Unlock()
		files, err := f.ExportContent()
		if err != nil {
			http.Error(w, "export failed", http.StatusInternalServerError)
			return
		}
		respond(w, map[string]any{"files": files})
	})
	return mux
}

func equal(a, b string) bool { return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1 }

func decode(w http.ResponseWriter, r *http.Request, target any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 8<<20)
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(target); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return false
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		http.Error(w, "expected one JSON value", http.StatusBadRequest)
		return false
	}
	return true
}

func respond(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(value)
}
