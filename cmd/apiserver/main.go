//go:build apiserver

package main

import (
	"api"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	cfg := api.Config{
		Addr:  envOr("API_ADDR", ":8787"),
		Root:  envOr("API_ROOT", defaultAPIRoot()),
		Token: os.Getenv("API_TOKEN"),
	}

	if err := os.MkdirAll(cfg.Root, 0755); err != nil {
		log.Fatal(err)
	}

	log.Printf("api listening on %s, root %s", cfg.Addr, cfg.Root)
	log.Fatal(api.NewServer(cfg).ListenAndServe())
}

func envOr(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func defaultAPIRoot() string {
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return "/home/asdf/api/prsnl.spc"
	}
	return filepath.Join(home, "api", "prsnl.spc")
}
