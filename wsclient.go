package main

import (
	"encoding/json"
	"net/http"
	"time"

	wsock "github.com/gorilla/websocket"
)

var wscResister = make(chan *WsClient, 8)
var wscUnregister = make(chan *WsClient, 8)
var wscSendTo = make(chan bool)

func wscHub() {
	var wscls = make(map[*WsClient]bool)
	// TODO need cleanup in case of exit?
	defer func() {
		glog.Warning("web-service client hub terminated")
	}()
	for {
		select {
		case c := <-wscResister:
			wscls[c] = true
		case c := <-wscUnregister:
			delete(wscls, c)
		case msg := <-wscSendTo:
			for c, _ := range wscls {
				c.events <- msg
			}
		}
	}
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

type WsClient struct {
	conn   *wsock.Conn
	events chan bool
}

var cmdReload = func() []byte {
	cmd := map[string]string{
		"cmd": "reload"}
	res, err := json.Marshal(cmd)
	if err != nil {
		panic("cannot prepare web-socket command 'reload'")
	}
	return res
}()

func (wsc *WsClient) talkTo() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		wsc.conn.Close()
	}()
	for {
		select {
		case _, ok := <-wsc.events:
			wsc.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				wsc.conn.WriteMessage(wsock.CloseMessage, []byte{})
				return
			}
			wr, err := wsc.conn.NextWriter(wsock.TextMessage)
			if err != nil {
				glog.Error("web-socket: cannot get writer:", err)
				return
			}
			wr.Write(cmdReload)

			// Add queued chat messages to the current websocket message.
			//			n := len(wsc.send)
			//			for i := 0; i < n; i++ {
			//				w.Write(newline)
			//				w.Write(<-c.send)
			//			}

			if err := wr.Close(); err != nil {
				glog.Error("web-socket: closing writer:", err)
				return
			}
		case <-ticker.C:
			wsc.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := wsc.conn.WriteMessage(wsock.PingMessage, nil); err != nil {
				glog.Error("web-socket: writing ping:", err)
				return
			}
		}
	}
}

func (wsc *WsClient) readFrom() {
	defer func() {
		glog.Debugf("drop web-socket client: %s", wsc.conn.RemoteAddr().String())
		wscUnregister <- wsc
		wsc.conn.Close()
	}()
	wsc.conn.SetReadLimit(maxMessageSize)
	wsc.conn.SetReadDeadline(time.Now().Add(pongWait))
	wsc.conn.SetPongHandler(func(string) error {
		wsc.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, msg, err := wsc.conn.ReadMessage()
		if err != nil {
			if wsock.IsUnexpectedCloseError(err, wsock.CloseGoingAway) {
				glog.Error("web-socket:", err)
			}
			break
		}
		glog.Noticef("web-socket incoming: [%s]", msg)
		jevt := make(map[string]interface{})
		err = json.Unmarshal(msg, jevt)
		if err != nil {
			glog.Error("cannot parse user event", err)
		} else {
			eventq <- bcEvent{esrcUsr, jevt}
		}
	}
}

var upgrader = wsock.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		glog.Error("cannot upgrade to seb-socket:", err)
		return
	}
	client := &WsClient{conn: conn, events: make(chan bool, 16)}
	wscResister <- client

	go client.talkTo()
	go client.readFrom()
	glog.Debugf("new web-socket client: %s", conn.RemoteAddr().String())
}
