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
	Data goxic.PhIdxs
}

func (tmpl *template) ServeHTTP(wr http.ResponseWriter, rq *http.Request) {
	var bt goxic.BounT
	tmpl.PrepareScreen(&bt)
	screen.EDState.Read(func() error {
		bt.Bind(goxic.Empty, tmpl.Data...)
		goxic.Must(bt.WriteTo(wr))
		return nil
	})
}
