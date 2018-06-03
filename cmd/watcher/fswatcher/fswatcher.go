// Copyright 2018 Christopher Cormack. All rights reserved.
// Use of this source code is governed by the MIT License
// that can be found in the LICENSE file.

package fswatcher

import (
	"github.com/cjcormack/file-watcher/pkg/debouncer"
	"github.com/fsnotify/fsnotify"
	"github.com/go-errors/errors"
	"log"
	"os"
	"path/filepath"
	"time"
)

type FsWatcher struct {
	watcher  *fsnotify.Watcher
	folders  map[string]*Folder
	isActive chan bool
	done     chan bool
	Events   chan *Event
	isClosed bool
}

func New() (*FsWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	fsWatcher := &FsWatcher{
		watcher:  watcher,
		isActive: make(chan bool),
		done:     make(chan bool),
		folders:  make(map[string]*Folder),
		Events:   make(chan *Event),
	}

	go fsWatcher.run()

	return fsWatcher, nil
}

func (this *FsWatcher) Start() {
	if this.isClosed {
		return
	}
	this.isActive <- true
}

func (this *FsWatcher) Stop() {
	if this.isClosed {
		return
	}
	this.isActive <- false
}

func (this *FsWatcher) run() {
	isActive := false

	for {
		select {
		case isActive = <-this.isActive:

		case event := <-this.watcher.Events:
			if isActive {
				err := this.handleEvent(event)
				if err != nil {
					log.Println(err.ErrorStack())
				}
			}
		case err := <-this.watcher.Errors:
			log.Printf("Received error from fsnotify (%s)", err)
		case <-this.done:
			this.isClosed = true
			return
		}
	}
}

func (this *FsWatcher) handleEvent(event fsnotify.Event) *errors.Error {
	if event.Op&(fsnotify.Create|fsnotify.Remove|fsnotify.Rename) == 0 {
		return nil
	}

	// Calculate the folder that this event was for
	parentFolder := filepath.Dir(event.Name)

	// We only pass on this event if is is for a folder that is registered as being watched.
	// It's unlikely that we'll receive such an event, but we don't want things to blow up
	folder, ok := this.folders[parentFolder]
	if !ok {
		log.Printf("Received event for parent folder that isn't being watched (%s)", parentFolder)
		return nil
	}

	folder.debouncer.Set(struct{}{})

	return nil
}

func (this *FsWatcher) AddFolder(path string) (*Folder, error) {
	folderToWatch, err := os.Stat(path)

	if err != nil {
		return nil, errors.Wrap(err, 0)
	}
	if !folderToWatch.IsDir() {
		return nil, errors.Errorf("'%s' is not a folder", path)
	}

	folder := &Folder{
		Name: path,
	}

	folder.debouncer = debouncer.New(debouncer.Opts{
		Duration: 200 * time.Millisecond,
		Callback: func(data interface{}) {
			this.Events <- &Event{
				Folder: folder,
			}
		},
	})

	this.folders[path] = folder

	this.watcher.Add(path)

	log.Printf("Watching '%s'", path)

	return folder, nil
}

func (this *FsWatcher) Close() error {
	if this.isClosed {
		return nil
	}

	this.done <- true
	err := this.watcher.Close()
	if err != nil {
		return errors.Wrap(err, 0)
	}

	return nil
}
