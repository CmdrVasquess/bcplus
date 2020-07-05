package travel

import (
	"net/http"

	"git.fractalqb.de/fractalqb/c4hgol"
	"git.fractalqb.de/fractalqb/goxic"
	"git.fractalqb.de/fractalqb/qbsllm"
	"github.com/CmdrVasquess/bcplus/internal/wapp"
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

func (tmpl *template) ServeHTTP(wr http.ResponseWriter, rq *http.Request) {
	if rq.Header.Get("Accept") == "application/json" {
	} else {
		if push, ok := wr.(http.Pusher); ok {
			opts := http.PushOptions{
				Header: http.Header{
					"Accept": []string{"application/json"},
				},
			}
			if err := push.Push("/"+screen.Key, &opts); err != nil {
				log.Errore(err)
			}
		}
		var bt goxic.BounT
		tmpl.PrepareScreen(&bt)
		screen.EDState.Read(func() error {
			goxic.Must(bt.WriteTo(wr))
			return nil
		})
	}
}
