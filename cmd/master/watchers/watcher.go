// Copyright 2018 Christopher Cormack. All rights reserved.
// Use of this source code is governed by the MIT License
// that can be found in the LICENSE file.

package watchers

import (
	"encoding/json"
	"github.com/cjcormack/file-watcher/pkg/message"
	"github.com/gorilla/websocket"
	"log"
	"sync"
)

type watcher struct {
	id            int
	conn          *websocket.Conn
	readErrors    int
	latestDetails *message.Folder
	lock          *sync.RWMutex
}

func (this *watcher) waitAndReceiveMessages() {
	for {
		mt, mb, err := this.conn.ReadMessage()

		if websocket.IsCloseError(err) || websocket.IsUnexpectedCloseError(err) {
			log.Printf("Connection closed (%s)", err)
			return
		}
		if err != nil {
			log.Printf("Read message error (%s)", err)
			this.readErrors++
			if this.readErrors >= 3 {
				log.Println("Too many concurrent read errors, killing the connection")
				return
			}
			continue
		}

		this.readErrors = 0

		if mt != websocket.TextMessage {
			log.Printf("Unexpected message type (%d)", mt)
			continue
		}

		var m message.Message
		err = json.Unmarshal(mb, &m)
		if err != nil {
			log.Printf("Unable to unmarshal JSON (%s)", err)
			continue
		}

		switch payload := m.Payload.(type) {
		case message.Folder:
			this.latestDetails = &payload
		default:
			log.Printf("Unhandled type of payload (%s)", err)

		}
	}
}

func (this *watcher) withLock(h func()) {
	this.lock.Lock()
	defer this.lock.Unlock()

	h()
}

func (this *watcher) withRLock(h func()) {
	this.lock.RLock()
	defer this.lock.RUnlock()

	h()
}
