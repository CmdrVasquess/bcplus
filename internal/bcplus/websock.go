package bcplus

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/CmdrVasquess/bcplus/internal/wapp"

	"git.fractalqb.de/fractalqb/qbsllm"

	wsock "github.com/gorilla/websocket"
)

var (
	webLog wsockWr
	webApp wsockWr
)

type wsockWr struct {
	lock   sync.Mutex
	wsConn *wsock.Conn
}

func (wlw *wsockWr) Set(c *wsock.Conn) bool {
	wlw.lock.Lock()
	defer wlw.lock.Unlock()
	if wlw.wsConn != nil {
		return false
	}
	wlw.wsConn = c
	return true
}

func (wlw *wsockWr) Write(p []byte) (n int, err error) {
	wlw.lock.Lock()
	defer wlw.lock.Unlock()
	if wlw.wsConn != nil {
		err = wlw.wsConn.WriteMessage(wsock.TextMessage, p)
		if err != nil {
			log.Warna("web socket `err`", err)
			return 0, err
		}
	}
	return len(p), nil
}

func (wlw *wsockWr) Close() {
	wlw.lock.Lock()
	defer wlw.lock.Unlock()
	if wlw.wsConn != nil {
		wlw.wsConn.Close()
		wlw.wsConn = nil
	}
}

func logWs(wr http.ResponseWriter, rq *http.Request) {
	if wsc, err := (&wsock.Upgrader{}).Upgrade(wr, rq, nil); err != nil {
		log.Errora("cannot upgrade to logging web-socket: `err`", err)
		return
	} else if !webLog.Set(wsc) {
		wsc.Close()
		log.Warna("rejected log `client`", wsc.RemoteAddr().String())
		return
	}
	defer webLog.Close()
	log.Infoa("new log `client`", webLog.wsConn.RemoteAddr().String())
	for webLog.wsConn != nil { // TODO lock
		_, _, err := webLog.wsConn.ReadMessage()
		if err != nil {
			log.Infoa("closed log `client` `because`",
				webLog.wsConn.RemoteAddr().String(),
				err)
			webLog.Close() // TODO Why isn't is enought to close in defer?
			break
		}
		log.Errora("incoming on log ws `from`", webLog.wsConn.RemoteAddr().String())
	}
}

type wsEvent struct {
	To  string
	Key string
	Cmd interface{}
}

func appWs(wr http.ResponseWriter, rq *http.Request) {
	wsc, err := (&wsock.Upgrader{}).Upgrade(wr, rq, nil)
	if err != nil {
		log.Errora("cannot upgrade to app web-socket: `err`", err)
		return
	} else if !webApp.Set(wsc) {
		wsc.Close()
		log.Warna("rejected app `client`", wsc.RemoteAddr().String())
		return
	}
	defer wsc.Close()
	log.Infoa("new app `client`", webApp.wsConn.RemoteAddr().String())
	for webApp.wsConn != nil {
		mty, mraw, err := webApp.wsConn.ReadMessage()
		if err != nil {
			log.Infoa("closed log `client` `because`",
				webApp.wsConn.RemoteAddr().String(),
				err)
			webApp.Close()
			break
		}
		switch mty {
		case wsock.TextMessage:
			evt := new(wsEvent)
			err := json.Unmarshal(mraw, evt)
			if err != nil {
				log.Errore(err)
			} else {
				if log.Logs(qbsllm.Ltrace) {
					log.Infof("WS: %s", string(mraw))
				}
				App.evtq <- evt
			}
		case wsock.BinaryMessage:
			log.Errora("ignore binary app event `from`",
				webApp.wsConn.RemoteAddr().String())
		}
	}
}

func doWsEvent(evt *wsEvent) {
	switch evt.To {
	case "screen":
		scr := wapp.Screens[evt.Key]
		if scr == nil {
			log.Errora("unkown `screen` as WS event target", evt.Key)
		} else {
			var msg []byte
			err := App.EDState.Read(func() (err error) {
				var data struct {
					Cmd  string
					Hdr  wapp.ScreenHdr
					Data interface{}
				}
				data.Cmd = "upd"
				data.Hdr.Set(App.EDState)
				data.Data = scr.Handler.Data()
				msg, err = json.Marshal(&data)
				return err
			})
			if err != nil {
				log.Errore(err)
			} else {
				webApp.Write(msg)
			}
		}
	default:
		log.Errora("unkonwn WS event `target`", evt.To)
	}
}
