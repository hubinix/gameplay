package http

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	//拿到连接了,开始维护吧
	//state     int
	//	queue     util.Queue
	//	notify    chan struct{}
	//	sendQueue util.Queue
	//	send      chan struct{}
	//	event     chan struct{}
	//	done      chan struct{}
	//	// The websocket connection.
	//	conn *websocket.Conn
	client := NewClientConnection(conn)

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.write()
	go client.read()
}
