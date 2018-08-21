package main

import (
	"encoding/json"
	"time"

	"git.fractalqb.de/fractalqb/ggja"
	l "git.fractalqb.de/fractalqb/qblog"
	"github.com/CmdrVasquess/BCplus/galaxy"
	eddn "github.com/CmdrVasquess/goEDDNc"
)

func ggjaFailLogErr(err error) {
	log.Log(l.Error, err)
}

func eddnTraceMsg(msg map[string]interface{}) {
	if logEddn.Logs(l.Trace) {
		trc, err := json.Marshal(msg)
		if err != nil {
			logEddn.Log(l.Error, err)
		} else {
			logEddn.Log(l.Trace, "EDDN << ", string(trc))
		}
	}
}

func eddnSendErr(err error, msg map[string]interface{}) {
	if err != nil {
		logEddn.Log(l.Warn, err)
		if logV {
			ej, _ := json.Marshal(msg)
			logEddn.Log(l.Debug, string(ej))
		}
	}
}

func eddnSendJournal(upld *eddn.Upload, ts time.Time, e ggja.GenObj, ssys *galaxy.System) {
	jump := eddn.NewMessage(eddn.Ts(ts))
	err := eddn.SetJournal(jump, e, ssys.Name,
		ssys.Coos[galaxy.Xk], ssys.Coos[galaxy.Yk], ssys.Coos[galaxy.Zk],
		true)
	if err != nil {
		logEddn.Log(l.Warn, err)
		return
	}
	logEddn.Logf(l.Debug, "send journal event to EDDN (dry: %t, test: %t)",
		upld.DryRun,
		upld.TestUrl)
	eddnTraceMsg(jump)
	err = upld.Send(eddn.Sjournal, jump)
	logEddn.Log(l.Trace, "done with journal event to EDDN")
	eddnSendErr(err, jump)
}

func eddnSendCommodities(upld *eddn.Upload, jStat map[string]interface{}) {
	stat := ggja.Obj{Bare: jStat, OnError: ggjaFailLogErr}
	tstr := stat.MStr("timestamp")
	if len(tstr) == 0 {
		return
	}
	logEddn.Logf(l.Debug, "send commodities to EDDN (dry: %t, test: %t)",
		upld.DryRun,
		upld.TestUrl)
	cmdt := eddn.NewMessage(tstr)
	err := eddn.SetCommoditiesJ(cmdt, jStat)
	if err != nil {
		log.Logf(l.Error, "eddn commodities: %s", err)
		return
	}
	eddnTraceMsg(cmdt)
	err = upld.Send(eddn.Scommodity, cmdt)
	logEddn.Log(l.Trace, "done with commodities to EDDN")
	eddnSendErr(err, cmdt)
}

func eddnSendShipyard(upld *eddn.Upload, jShy map[string]interface{}) {
	shy := ggja.Obj{Bare: jShy, OnError: ggjaFailLogErr}
	tstr := shy.MStr("timestamp")
	if len(tstr) == 0 {
		return
	}
	logEddn.Logf(l.Debug, "send shipyard to EDDN (dry: %t, test: %t)",
		upld.DryRun,
		upld.TestUrl)
	cmdt := eddn.NewMessage(tstr)
	err := eddn.SetShipyardJ(cmdt, jShy)
	if err != nil {
		log.Logf(l.Error, "eddn shipyard: %s", err)
		return
	}
	eddnTraceMsg(cmdt)
	err = upld.Send(eddn.Sshipyard, cmdt)
	logEddn.Log(l.Trace, "done with shipyard to EDDN")
	eddnSendErr(err, cmdt)
}
