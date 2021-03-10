package http

import (
	"github.com/gorilla/websocket"
	"github.com/hubinix/gameplay/pkg/util"
	"log"
	"time"
)

// Client is a middleman between the websocket connection and the hub.
type ClientConnection struct {
	queue     *util.Queue
	notify    chan struct{}
	sendQueue *util.Queue
	send      chan struct{}
	event     chan struct{}
	done      chan struct{}
	// The websocket connection.
	conn *websocket.Conn
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

func NewClientConnection(conn *websocket.Conn) *ClientConnection {
	return &ClientConnection{
		queue:     util.NewQueue(maxMessageSize),
		notify:    make(chan struct{}, 1),
		sendQueue: util.NewQueue(maxMessageSize),
		send:      make(chan struct{}, 1),
		event:     make(chan struct{}, 1),
		done:      make(chan struct{}, 1),
		// The websocket connection.
		conn: conn,
	}
}

func (c *ClientConnection) read() {
	defer func() {
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		//扔到队列里面
		c.queue.Put(message)

	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *ClientConnection) write() {

	defer func() {
		c.conn.Close()
	}()
	//tick
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			w, err := c.conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				return
			}
			msg, ok := c.sendQueue.Get()
			for ok {
				data, ok := msg.([]byte)
				if ok {
					err := c.conn.WriteMessage(websocket.BinaryMessage, []byte(data))
					if err != nil {
						log.Println("write:", err)

						return
					}
				}
				msg, ok = c.sendQueue.Get()
			}

			// Add queued chat messages to the current websocket message.

			if err := w.Close(); err != nil {
				return
			}
		case <-c.done:
			// The hub closed the channel.
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
