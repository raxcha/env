package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"env/api"
	"env/filesystem"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	root := flag.String("root", filepath.Join(home, "prsnlspc"), "filesystem directory")
	addr := flag.String("addr", "127.0.0.1:8081", "backend listen address")
	flag.Parse()
	f, err := filesystem.Open(*root)
	if err != nil {
		log.Fatal(err)
	}
	server := &http.Server{Addr: *addr, Handler: api.New(f, os.Getenv("API_USERNAME"), os.Getenv("API_PASSWORD")), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	log.Printf("API listening on %s; filesystem: %s", *addr, *root)
	log.Fatal(server.ListenAndServe())
}
