package travel

import (
	"net/http"
	"time"

	"git.fractalqb.de/fractalqb/c4hgol"
	"git.fractalqb.de/fractalqb/goxic"
	"git.fractalqb.de/fractalqb/qbsllm"
	"github.com/CmdrVasquess/bcplus/internal/wapp"
	"github.com/CmdrVasquess/goedx"
)

var (
	log    = qbsllm.New(qbsllm.Lnormal, "travel", nil, nil)
	LogCfg = c4hgol.Config(qbsllm.NewConfig(log))
)

func init() {
	wapp.AddScreen(&screen, LogCfg)
	tmpl.BCpScreen = &screen
}

var (
	tmpl   template
	screen = wapp.Screen{
		Key: "travel",
		Tab: "Travel",
		// Title:   "Travel", same as Tab
		Handler: &tmpl,
	}
)

type template struct {
	wapp.ScreenTmpl
}

type Jump struct {
	Name string
	Time time.Time
	Coos *goedx.SysCoos
}

type Data struct {
	JumpHist []Jump
}

func (data *Data) Set(ed *goedx.EDState) {
	if cmdr := screen.Ext.EDState.Cmdr; cmdr != nil {
		gxy := screen.Ext.Galaxy
		for _, jump := range cmdr.JumpHist {
			sys, _ := gxy.EdgxSystem(jump.SysAddr, "", nil, time.Time{})
			data.JumpHist = append(data.JumpHist, Jump{
				Name: sys.Name,
				Time: jump.Arrive,
				Coos: &sys.Coos,
			})
		}
	} else {
		data.JumpHist = []Jump{}
	}
}

func (tmpl *template) ServeHTTP(wr http.ResponseWriter, rq *http.Request) {
	if wapp.DataRequest(rq) {
		var data struct {
			Hdr wapp.ScreenHdr
			Data
		}
		err := screen.Ext.EDState.Read(func() error {
			data.Hdr.Set(screen.Ext.EDState)
			data.Set(screen.Ext.EDState)
			return wapp.DataResponse(wr, &data)
		})
		if err != nil {
			panic(err)
		}
	} else {
		// TODO does http push work with XMLHttpRequest() ???
		// if push, ok := wr.(http.Pusher); ok {
		// 	opts := http.PushOptions{
		// 		Header: http.Header{
		// 			"Accept": []string{"application/json"},
		// 		},
		// 	}
		// 	if err := push.Push("/"+screen.Key, &opts); err != nil {
		// 		log.Errore(err)
		// 	}
		// }
		var bt goxic.BounT
		tmpl.PrepareScreen(&bt)
		screen.Ext.EDState.Read(func() error {
			goxic.Must(bt.WriteTo(wr))
			return nil
		})
	}
}
