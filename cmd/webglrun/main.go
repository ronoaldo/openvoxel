package main

import (
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

func main() {
	log.Print("Watching for file changes ... ")
	go watchForChanges()

	_, err := exec.Command("xdg-open", "http://localhost:8080/").CombinedOutput()
	log.Printf("Launching browser (err=%v)", err)

	log.Print("Starting server for wasmrun ...")
	http.Handle("/", http.FileServer(http.Dir("./")))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func watchForChanges() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	lastBuild := time.Now()
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Printf("Event: %#v", event)
				if strings.HasSuffix(event.Name, ".go") {
					if time.Since(lastBuild) < 100*time.Millisecond {
						log.Printf("Not rebuilding since lastBuild is %v ago", time.Since(lastBuild))
						continue
					}

					b, err := exec.Command("scripts/make.sh", "js", "wasm").CombinedOutput()
					if err != nil {
						log.Printf("Error: %v", err)
					} else {
						log.Printf("Build output: %v", string(b))
					}
					lastBuild = time.Now()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("Error: %v", err)
			}
		}
	}()

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	log.Printf("Watching for changes at '%v'", wd)
	filepath.Walk(wd, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() && !strings.Contains(path, "/.git") && !strings.Contains(path, "/build") {
			log.Printf("> Added %v", path)
			return watcher.Add(path)
		}
		return nil
	})
}
