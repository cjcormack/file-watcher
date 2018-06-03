// Copyright 2018 Christopher Cormack. All rights reserved.
// Use of this source code is governed by the MIT License
// that can be found in the LICENSE file.

package main

import (
	"flag"
	"github.com/cjcormack/file-watcher/cmd/master/middleware"
	"github.com/cjcormack/file-watcher/cmd/master/watchers"
	"goji.io"
	"goji.io/pat"
	"log"
	"net/http"
)

var watcherAddr = flag.String("watcherAddr", "localhost:8080", "http service address")
var publicAddr = flag.String("publicAddr", "localhost:8090", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)

	watcherState := watchers.NewWatcherState()

	watchersMux := goji.NewMux()
	watchersMux.HandleFunc(pat.Get("/watchers/notifications"), watcherState.WatcherRoute)

	publicMux := goji.NewMux()
	publicMux.Use(middleware.Logger)
	publicMux.HandleFunc(pat.Get("/files"), watcherState.CurrentFilesRoute)

	log.Printf("Starting watchers HTTP server on %s", *watcherAddr)
	go func() {
		log.Fatal(http.ListenAndServe(*watcherAddr, watchersMux))
	}()

	log.Printf("Starting public HTTP server on %s", *publicAddr)
	log.Fatal(http.ListenAndServe(*publicAddr, publicMux))
}
