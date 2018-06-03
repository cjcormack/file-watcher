// Copyright 2018 Christopher Cormack. All rights reserved.
// Use of this source code is governed by the MIT License
// that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"

	"github.com/cjcormack/file-watcher/cmd/watcher/fswatcher"
	"github.com/cjcormack/file-watcher/cmd/watcher/socket"
	"github.com/cjcormack/file-watcher/pkg/message"
	"os"
	"os/signal"
)

var addr = flag.String("addr", "localhost:8080", "http service address")
var path = flag.String("folder", "", "Path to the folder to watch")

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)

	watcher, err := fswatcher.New()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	watchedFolder, err := watcher.AddFolder(*path)
	if err != nil {
		log.Fatal(err)
	}

	conn := socket.New(*addr, "/watchers/notifications")
	conn.SetConnectHandler(connectionHandler(watcher, conn, []*fswatcher.Folder{watchedFolder}))
	conn.SetDisconnectHandler(disconnectionHandler(watcher))

	conn.SetDisconnectHandler(func() {
		watcher.Stop()
	})

	conn.Connect()
	defer conn.Close()

	run(watcher, conn)
}

func connectionHandler(watcher *fswatcher.FsWatcher, conn *socket.Connection, folders []*fswatcher.Folder) func() {
	return func() {
		watcher.Start()

		for _, folder := range folders {
			sendContentsForFolder(conn, folder)
		}
	}
}

func disconnectionHandler(watcher *fswatcher.FsWatcher) func() {
	return func() {
		watcher.Stop()
	}
}

func run(watcher *fswatcher.FsWatcher, conn *socket.Connection) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	for {
		select {
		case event := <-watcher.Events:
			sendContentsForFolder(conn, event.Folder)

		case msg := <-conn.Messages:
			log.Printf("Received socket message (%v)", msg)

		case <-sigChan:
			return
		}
	}
}

func sendContentsForFolder(conn *socket.Connection, folder *fswatcher.Folder) {
	currentFiles, err := folder.GetCurrentFiles()
	if err != nil {
		log.Printf("Could not get the current files (%v)", err)
		return
	}

	msg := &message.Message{
		Type: message.FolderContents,
		Payload: message.Folder{
			Name:  folder.Name,
			Files: currentFiles,
		},
	}

	conn.SendMessage(msg)
}
