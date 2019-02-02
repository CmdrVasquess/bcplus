package webui

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/CmdrVasquess/BCplus/common"
	wsock "github.com/gorilla/websocket"
)

const WsLoadCmd = "load"

type WsCommand struct {
	Cmd string
}

type WsCmdLoad struct {
	WsCommand
	Url string `json:",omitempty"`
}

var wscResister = make(chan *WsClient, 8)
var wscUnregister = make(chan *WsClient, 8)
var wscSendTo = make(chan interface{}, 8)

func wscHub() {
	log.Infos("running WebSocket-client hub…")
	var wscls = make(map[*WsClient]bool)
	// TODO need cleanup in case of exit?
	defer func() {
		log.Warns("web-service client hub terminated")
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
	events chan interface{}
}

var wscReload = map[string]string{"cmd": "reload"}

func (wsc *WsClient) talkTo() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		wsc.conn.Close()
	}()
	for {
		select {
		case evt, ok := <-wsc.events:
			msg, err := json.Marshal(evt)
			if err != nil {
				log.Errora("web-socket: cannot marshal message: `err`", err)
				return
			}
			wsc.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				wsc.conn.WriteMessage(wsock.CloseMessage, []byte{})
				return
			}
			wr, err := wsc.conn.NextWriter(wsock.TextMessage)
			if err != nil {
				log.Errora("web-socket: cannot get writer: `err`", err)
				return
			}
			wr.Write(msg)

			// Add queued chat messages to the current websocket message.
			//			n := len(wsc.send)
			//			for i := 0; i < n; i++ {
			//				w.Write(newline)
			//				w.Write(<-c.send)
			//			}

			if err := wr.Close(); err != nil {
				log.Errora("web-socket: closing writer: `err`", err)
				return
			}
		case <-ticker.C:
			wsc.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := wsc.conn.WriteMessage(wsock.PingMessage, nil); err != nil {
				log.Errora("web-socket: writing ping: `err`", err)
				return
			}
		}
	}
}

func (wsc *WsClient) readFrom() {
	defer func() {
		log.Debuga("drop web-socket `client`", wsc.conn.RemoteAddr().String())
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
				log.Errora("web-socket: `err`", err)
			}
			break
		}
		log.Tracea("web-socket incoming `msg`", msg)
		jevt := make(map[string]interface{})
		err = json.Unmarshal(msg, &jevt)
		if err != nil {
			log.Errora("cannot parse user event: `err`", err)
		} else {
			log.Tracea("dispatch user event to main event-q: `evt`", jevt)
			theBCpQ <- common.BCpEvent{common.BCpEvtSrcWUI, jevt}
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
		log.Errora("cannot upgrade to web-socket: `err`", err)
		return
	}
	client := &WsClient{conn: conn, events: make(chan interface{}, 16)}
	wscResister <- client

	go client.talkTo()
	go client.readFrom()
	log.Debuga("new web-socket `client`", conn.RemoteAddr().String())
}
