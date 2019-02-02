package core

import (
	"encoding/json"
	"time"

	"git.fractalqb.de/fractalqb/ggja"
	log "git.fractalqb.de/fractalqb/qblog"
	"github.com/CmdrVasquess/BCplus/galaxy"
	eddn "github.com/CmdrVasquess/goEDDNc"
)

const (
	flagEddnDefault = "off"
	flagEddnOff     = "off"
	eddnTimeout     = 8 * time.Second
)

var eddnMode string

func ggjaFailLogErr(err error) {
	logEddn.Errora("`err`", err)
}

func eddnTraceMsg(msg map[string]interface{}) {
	if logEddn.Logs(log.Ltrace) {
		trc, err := json.Marshal(msg)
		if err != nil {
			logEddn.Errora("`err`", err)
		} else {
			logEddn.Tracea("EDDN << `msg`", string(trc))
		}
	}
}

func eddnSendErr(err error, msg map[string]interface{}) {
	if err != nil {
		logEddn.Warna("`err`", err)
		if LogV {
			ej, _ := json.Marshal(msg)
			logEddn.Debug(log.Str(string(ej)))
		}
	}
}

func eddnSendJournal(upld *eddn.Upload, ts time.Time, e ggja.Obj, ssys *galaxy.System) {
	if eddnMode == flagEddnOff || jevtSpooling {
		return
	}
	go func() {
		jump := eddn.NewMessage(eddn.Ts(ts))
		err := eddn.SetJournal(jump, e.Bare, ssys.Name,
			ssys.Coos[galaxy.Xk], ssys.Coos[galaxy.Yk], ssys.Coos[galaxy.Zk],
			true)
		if err != nil {
			logEddn.Warna("`err`", err)
			return
		}
		logEddn.Debuga("send journal event to EDDN (`dry`, `test`)",
			upld.DryRun,
			upld.TestUrl)
		eddnTraceMsg(jump)
		err = upld.Send(eddn.Sjournal, jump)
		logEddn.Trace(log.Str("done with journal event to EDDN"))
		eddnSendErr(err, jump)
	}()
}

func eddnSendCommodities(upld *eddn.Upload, jStat map[string]interface{}) {
	stat := ggja.Obj{Bare: jStat, OnError: ggjaFailLogErr}
	tstr := stat.MStr("timestamp")
	if len(tstr) == 0 {
		return
	}
	logEddn.Debuga("send commodities to EDDN (`dry`, `test`)",
		upld.DryRun,
		upld.TestUrl)
	cmdt := eddn.NewMessage(tstr)
	err := eddn.SetCommoditiesJ(cmdt, jStat)
	if err != nil {
		lgr.Errora("eddn commodities: `err`", err)
		return
	}
	eddnTraceMsg(cmdt)
	err = upld.Send(eddn.Scommodity, cmdt)
	logEddn.Trace(log.Str("done with commodities to EDDN"))
	eddnSendErr(err, cmdt)
}

func eddnSendShipyard(upld *eddn.Upload, jShy map[string]interface{}) {
	shy := ggja.Obj{Bare: jShy, OnError: ggjaFailLogErr}
	tstr := shy.MStr("timestamp")
	if len(tstr) == 0 {
		return
	}
	logEddn.Debuga("send shipyard to EDDN (`dry`, `test`)",
		upld.DryRun,
		upld.TestUrl)
	shpy := eddn.NewMessage(tstr)
	err := eddn.SetShipyardJ(shpy, jShy)
	if err != nil {
		lgr.Errora("eddn shipyard: `err`", err)
		return
	}
	eddnTraceMsg(shpy)
	err = upld.Send(eddn.Sshipyard, shpy)
	logEddn.Trace(log.Str("done with shipyard to EDDN"))
	eddnSendErr(err, shpy)
}

func eddnSendOutfitting(upld *eddn.Upload, jOtf map[string]interface{}) {
	otf := ggja.Obj{Bare: jOtf, OnError: ggjaFailLogErr}
	tstr := otf.MStr("timestamp")
	if len(tstr) == 0 {
		return
	}
	logEddn.Debuga("send outfitting to EDDN (`dry`, `test`)",
		upld.DryRun,
		upld.TestUrl)
	otft := eddn.NewMessage(tstr)
	err := eddn.SetOutfittingJ(otft, jOtf)
	if err != nil {
		lgr.Errora("eddn outfitting: `err`", err)
		return
	}
	eddnTraceMsg(otft)
	err = upld.Send(eddn.Scommodity, otft)
	logEddn.Trace(log.Str("done with outfitting to EDDN"))
	eddnSendErr(err, otft)
}
