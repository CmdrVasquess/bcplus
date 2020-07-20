package wapp

import (
	"encoding/json"
	"html/template"
	"net/http"
	"reflect"

	"git.fractalqb.de/fractalqb/goxic/content"

	"git.fractalqb.de/fractalqb/c4hgol"
	"git.fractalqb.de/fractalqb/goxic"
	"git.fractalqb.de/fractalqb/qbsllm"
	"github.com/CmdrVasquess/goedx"
)

var (
	log    = qbsllm.New(qbsllm.Lnormal, "wapp", nil, nil)
	LogCfg = c4hgol.Config(qbsllm.NewConfig(log))

	jsonNull = []byte("null")
	tabBar   []byte
)

type ScreenTmpl struct {
	*goxic.Template
	BCpScreen  *Screen
	ScreenTabs goxic.PhIdxs
}

type Handler interface {
	http.Handler
	// Request the data for the change chg to be sent to the web UI. When
	// chg == 0 return the data to initialize the web UI.
	Data(chg goedx.Change) interface{}
}

func (st *ScreenTmpl) PrepareScreen(bt *goxic.BounT) {
	st.Template.NewBounT(bt)
	bt.Bind(content.Data(tabBar), st.ScreenTabs...)
}

type Screen struct {
	Key     string
	Tab     string
	Handler Handler          `json:"-"`
	Ext     *goedx.Extension `json:"-"`
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

var Screens = make(map[string]*Screen)

func InitTabsBar(order []string) {
	type tab struct{ Key, Tab string }
	bar := []tab{}
	have := make(map[string]bool)
	for _, t := range order {
		scrn := Screens[t]
		if scrn == nil {
			log.Warna("Unkown `tab` to init tab bar", t)
			continue
		}
		bar = append(bar, tab{Key: scrn.Key, Tab: template.JSEscapeString(scrn.Tab)})
		have[t] = true
	}
	for _, scrn := range Screens {
		if have[scrn.Key] {
			continue
		}
		bar = append(bar, tab{Key: scrn.Key, Tab: template.JSEscapeString(scrn.Tab)})
	}
	data, err := json.Marshal(bar)
	if err != nil {
		log.Fatale(err)
	}
	tabBar = data
}

// func DataRequest(rq *http.Request) bool {
// 	return rq.Header.Get("Accept") == "application/json"
// }

// func DataResponse(wr http.ResponseWriter, data interface{}) error {
// 	wr.Header().Set("Content-Type", "application/json")
// 	enc := json.NewEncoder(wr)
// 	return enc.Encode(&data)
// }

type ScreenHdr struct {
	Cmdr string
	Ship struct {
		Type  string
		Ident string
		Name  string
		Jump  float32
		Range float32
		Cargo int
	}
	Loc goedx.JSONLocation
}

func NewScreenHdr(ed *goedx.EDState) *ScreenHdr {
	res := new(ScreenHdr)
	res.Set(ed)
	return res
}

func (hdr *ScreenHdr) Set(ed *goedx.EDState) {
	if cmdr := ed.Cmdr; cmdr == nil {
		hdr.Cmdr = "<offline>"
		hdr.Ship.Type = "<type>"
		hdr.Ship.Ident = "<ident>"
		hdr.Ship.Name = "<name>"
		hdr.Loc.Location = nil
	} else {
		hdr.Cmdr = cmdr.Name
		ship := cmdr.GetShip(cmdr.ShipID)
		hdr.Ship.Type = ship.Type
		hdr.Ship.Ident = string(ship.Ident)
		hdr.Ship.Name = string(ship.Name)
		hdr.Ship.Jump = float32(ship.MaxJump)
		hdr.Ship.Range = float32(ship.MaxRange)
		hdr.Ship.Cargo = int(ship.Cargo)
		hdr.Loc = cmdr.At
	}
}
