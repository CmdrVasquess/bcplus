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

const Key = "travel"

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
	var bt goxic.BounT
	tmpl.PrepareScreen(&bt)
	screen.Ext.EDState.Read(func() error {
		goxic.Must(bt.WriteTo(wr))
		return nil
	})
}

func (tmpl *template) Data() interface{} {
	data := new(Data)
	data.Set(screen.Ext.EDState)
	return data
}
