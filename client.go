package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512

	clientOutgoingChannelSize = 256
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	//	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan *Message

	hub *hub

	wg sync.WaitGroup
}

func (c *Client) runRead() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			logrus.WithError(errors.WithStack(err)).Error("failed ReadMessage")
			return
		}

		fmt.Print(message)

		//c.hub.broadcast <- message
	}
}
func (c *Client) runWrite() {
	ping := time.NewTicker(pingPeriod)
	defer func() {
		c.conn.Close()
		ping.Stop()
		c.wg.Done()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// client channel was closed
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				logrus.WithError(errors.WithStack(err)).Error("failed create writer")
				return
			}

			data := msg.toData()
			w.Write(data)
			w.Close()
		case <-ping.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logrus.WithError(errors.WithStack(err)).Error("failed write PingMessage")
				return
			}
		}
	}
}

func (c *Client) WaitClosed() {
	c.wg.Wait()
}

func NewClient(w http.ResponseWriter, r *http.Request, h *hub) *Client {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.WithError(errors.WithStack(err)).Error("upgrade failed")
	}
	result := &Client{
		conn: conn,
		send: make(chan *Message, clientOutgoingChannelSize),
		hub:  h,
	}

	result.wg.Add(1)
	go result.runRead()
	result.wg.Add(1)
	go result.runWrite()

	result.hub.register <- result

	return result
}
