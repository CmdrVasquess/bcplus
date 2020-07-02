package wapp

import (
	"net/http"
	"reflect"

	"git.fractalqb.de/fractalqb/c4hgol"
	"git.fractalqb.de/fractalqb/goxic"
	"git.fractalqb.de/fractalqb/qbsllm"
)

var (
	log    = qbsllm.New(qbsllm.Lnormal, "wapp", nil, nil)
	LogCfg = c4hgol.Config(qbsllm.NewConfig(log))
)

type ScreenTmpl struct {
	http.Handler
	*goxic.Template
	Theme     goxic.PhIdxs
	ActiveTab goxic.PhIdxs
	InitHdr   goxic.PhIdxs
}

type Screen struct {
	Key      string
	Tab      string
	Title    string
	Template interface{}
}

func AddScreen(s *Screen) {
	if s.Key == "" {
		log.Fatala("empty `screen key` on `template`",
			s.Key,
			reflect.TypeOf(s.Template))
	}
	if tmp := Screens[s.Key]; tmp != nil {
		log.Fatala("duplicate `screen key` on `template 1` and `template 2`",
			s.Key,
			reflect.TypeOf(tmp.Template),
			reflect.TypeOf(s.Template))
	}
	Screens[s.Key] = s
}

var (
	Screens = make(map[string]*Screen)
)
