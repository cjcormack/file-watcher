// Copyright 2018 Christopher Cormack. All rights reserved.
// Use of this source code is governed by the MIT License
// that can be found in the LICENSE file.

package socket

import (
	"encoding/json"
	"github.com/cjcormack/file-watcher/pkg/message"
	"github.com/go-errors/errors"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"time"
)

type Connection struct {
	u    url.URL
	conn *websocket.Conn

	// Handlers for status changes
	connectHandler    func()
	disconnectHandler func()

	// Channels for communicating messages and errors from the connection
	Messages chan *message.Message

	// Channels for communicating with and controlling the go routine
	reconnect               chan bool
	connectOrPingTickerChan <-chan time.Time
	messageQueue            chan *message.Message
}

type Status int

func New(addr, path string) *Connection {
	u := url.URL{Scheme: "ws", Host: addr, Path: path}

	connection := &Connection{
		u: u,

		Messages: make(chan *message.Message),

		reconnect:    make(chan bool, 1),
		messageQueue: make(chan *message.Message, 10),
	}

	go connection.run()

	return connection
}

func (this *Connection) Connect() {
	if this.connectOrPingTickerChan == nil {
		this.connectOrPingTickerChan = time.NewTicker(2 * time.Second).C
		this.reconnect <- true
	}
}

func (this *Connection) SetConnectHandler(handler func()) {
	this.connectHandler = handler
}

func (this *Connection) SetDisconnectHandler(handler func()) {
	this.disconnectHandler = handler
}

func (this *Connection) run() {
	for {
		select {
		case <-this.reconnect:
			err := this.doClose()
			if err != nil {
				log.Printf("Could not close existing connection (%s)", err)
			}

			this.attemptConnect()
		case <-this.connectOrPingTickerChan:
			if this.conn == nil {
				this.attemptConnect()
			} else {
				this.conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
				if err := this.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					this.reconnectDueToError()
				}
			}
		case msg := <-this.messageQueue:
			// ignore any messages while we're not connected
			if this.conn != nil {
				err := this.doSendMessage(msg)
				if err != nil {
					log.Printf("Error sending message (%v)", err)
				}
			}
		}
	}
}

func (this *Connection) attemptConnect() {
	if this.conn != nil {
		// already connected
		return
	}

	log.Printf("Attempting to connect to %s", this.u.String())

	conn, _, err := websocket.DefaultDialer.Dial(this.u.String(), nil)
	if err != nil {
		log.Printf("Unable to connect (%s)", err)
	} else {
		log.Println("Connected")
		this.conn = conn

		this.conn.SetCloseHandler(func(code int, text string) error {
			log.Printf("Disconnected (%d): %s", code, text)

			msg := websocket.FormatCloseMessage(code, "")
			this.conn.WriteControl(websocket.CloseMessage, msg, time.Now().Add(time.Second))

			this.conn = nil
			this.disconnectHandler()
			this.reconnect <- true

			return nil
		})

		this.connectHandler()
	}
}

func (this *Connection) doClose() *errors.Error {
	if this.conn != nil {
		err := this.conn.Close()

		this.conn = nil // we always nil the existing conn, even on failure

		if err != nil {
			return errors.Wrap(err, 0)
		}

		log.Println("Connection closed")
	}

	return nil
}

func (this *Connection) reconnectDueToError() {
	this.disconnectHandler()
	this.reconnect <- true
}

func (this *Connection) doSendMessage(msg *message.Message) error {
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		return errors.Wrap(err, 0)
	}

	err = this.conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
	if err != nil {
		return errors.Wrap(err, 0)
	}

	err = this.conn.WriteMessage(websocket.TextMessage, jsonBytes)

	if err != nil {
		this.reconnectDueToError()
		// Something went wrong sending the message
		return errors.Wrap(err, 0)
	}

	return nil
}

func (this *Connection) SendMessage(msg *message.Message) {
	this.messageQueue <- msg
}

func (this *Connection) Close() error {
	err := this.doClose()

	if err != nil {
		return err
	}
	return nil
}
