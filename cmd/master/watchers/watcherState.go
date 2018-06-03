// Copyright 2018 Christopher Cormack. All rights reserved.
// Use of this source code is governed by the MIT License
// that can be found in the LICENSE file.

package watchers

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sort"
	"sync"
)

type WatcherState struct {
	watchers map[int]*watcher

	lastId int
	lock   *sync.RWMutex
}

func NewWatcherState() *WatcherState {
	return &WatcherState{
		watchers: make(map[int]*watcher),
		lock:     &sync.RWMutex{},
	}
}

var upgrader = websocket.Upgrader{} // use default options

func (this *WatcherState) withLock(h func()) {
	this.lock.Lock()
	defer this.lock.Unlock()

	h()
}

func (this *WatcherState) withRLock(h func()) {
	this.lock.RLock()
	defer this.lock.RUnlock()

	h()
}

func (this *WatcherState) WatcherRoute(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Unable to upgrade connection (%s)", err)
	}
	defer c.Close()

	watcher := &watcher{
		conn: c,
		lock: &sync.RWMutex{},
	}

	this.withLock(func() {
		id := this.lastId + 1
		this.lastId = id

		watcher.id = id

		this.watchers[id] = watcher
	})

	log.Printf("Watcher connected (%d)", watcher.id)

	defer func() {
		this.withLock(func() {
			delete(this.watchers, watcher.id)
		})
		log.Printf("Watcher disconnected (%d)", watcher.id)
	}()

	watcher.waitAndReceiveMessages()
}

type values struct {
	Files []file `json:"files"`
}

type file struct {
	FileName string `json:"filename"`
}

func (this *WatcherState) CurrentFilesRoute(w http.ResponseWriter, r *http.Request) {
	var allFileNames []string

	this.withRLock(func() {
		for _, watcher := range this.watchers {
			watcher.withRLock(func() {
				if watcher.latestDetails != nil {
					allFileNames = append(allFileNames, watcher.latestDetails.Files...)
				}
			})
		}
	})

	allFileNames = dedupe(allFileNames)
	sort.Strings(allFileNames)

	resp := values{
		Files: []file{},
	}

	for _, fileName := range allFileNames {
		resp.Files = append(resp.Files, file{fileName})
	}

	enc := json.NewEncoder(w)
	err := enc.Encode(resp)
	if err != nil {
		http.Error(w, "Error when creating response", http.StatusInternalServerError)
		log.Println("Error when creating response (%s)", err)
		return
	}
}

func dedupe(s []string) []string {
	seen := make(map[string]bool, len(s))
	deduped := make([]string, 0, len(s))

	for _, v := range s {
		if !seen[v] {
			seen[v] = true
			deduped = append(deduped, v)
		}
	}

	return deduped
}
