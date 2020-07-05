package wapp

import (
	"net/http"
	"reflect"

	"git.fractalqb.de/fractalqb/c4hgol"
	"git.fractalqb.de/fractalqb/goxic"
	"git.fractalqb.de/fractalqb/qbsllm"
	"github.com/CmdrVasquess/goedx"
)

var (
	log    = qbsllm.New(qbsllm.Lnormal, "wapp", nil, nil)
	LogCfg = c4hgol.Config(qbsllm.NewConfig(log))
)

type ScreenTmpl struct {
	*goxic.Template
	BCpScreen *Screen
}

var jsonNull = []byte("null")

func (st *ScreenTmpl) PrepareScreen(bt *goxic.BounT) {
	st.Template.NewBounT(bt)
	// if st.BCpScreen.EDState.Cmdr == nil {
	// 	bt.Bind(content.Data(jsonNull), st.InitHdr...)
	// } else {
	// 	bt.Bind(content.Json{V: st.BCpScreen.EDState.Cmdr}, st.InitHdr...)
	// }
}

type Screen struct {
	Key     string
	Tab     string
	Title   string
	Handler http.Handler
	EDState *goedx.EDState
}

func AddScreen(s *Screen, logCfg c4hgol.Configurer) {
	if s.Key == "" {
		log.Fatala("empty `screen key` on `template`",
			s.Key,
			reflect.TypeOf(s.Handler))
	}
	if tmp := Screens[s.Key]; tmp != nil {
		log.Fatala("duplicate `screen key` on `template 1` and `template 2`",
			s.Key,
			reflect.TypeOf(tmp.Handler),
			reflect.TypeOf(s.Handler))
	}
	Screens[s.Key] = s
	c4hgol.Config(LogCfg, logCfg)
}

var (
	Screens = make(map[string]*Screen)
)
