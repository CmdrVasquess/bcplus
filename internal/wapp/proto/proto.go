package proto

import (
	"net/http"

	"git.fractalqb.de/fractalqb/goxic"

	"git.fractalqb.de/fractalqb/c4hgol"
	"git.fractalqb.de/fractalqb/qbsllm"
	"github.com/CmdrVasquess/bcplus/internal/wapp"
)

const Key = "proto"

var (
	log    = qbsllm.New(qbsllm.Lnormal, Key, nil, nil)
	LogCfg = c4hgol.Config(qbsllm.NewConfig(log))
)

func init() {
	wapp.AddScreen(&screen, LogCfg)
	tmpl.BCpScreen = &screen
}

var (
	tmpl   template
	screen = wapp.Screen{
		Key:     Key,
		Handler: &tmpl,
	}
)

type template struct {
	wapp.ScreenTmpl
}

func (tmpl *template) ServeHTTP(wr http.ResponseWriter, rq *http.Request) {
	var bt goxic.BounT
	tmpl.NewBounT(&bt)
	tmpl.PrepareScreen(&bt)
	goxic.Must(bt.WriteTo(wr))
}

func (tmpl *template) Data() interface{} {
	return "proto data"
}
