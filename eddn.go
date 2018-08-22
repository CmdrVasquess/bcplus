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
	log.Log(l.Lerror, err)
}

func eddnTraceMsg(msg map[string]interface{}) {
	if logEddn.Logs(l.Ltrace) {
		trc, err := json.Marshal(msg)
		if err != nil {
			logEddn.Log(l.Lerror, err)
		} else {
			logEddn.Log(l.Ltrace, "EDDN << ", string(trc))
		}
	}
}

func eddnSendErr(err error, msg map[string]interface{}) {
	if err != nil {
		logEddn.Log(l.Lwarn, err)
		if logV {
			ej, _ := json.Marshal(msg)
			logEddn.Log(l.Ldebug, string(ej))
		}
	}
}

func eddnSendJournal(upld *eddn.Upload, ts time.Time, e ggja.GenObj, ssys *galaxy.System) {
	jump := eddn.NewMessage(eddn.Ts(ts))
	err := eddn.SetJournal(jump, e, ssys.Name,
		ssys.Coos[galaxy.Xk], ssys.Coos[galaxy.Yk], ssys.Coos[galaxy.Zk],
		true)
	if err != nil {
		logEddn.Log(l.Lwarn, err)
		return
	}
	logEddn.Logf(l.Ldebug, "send journal event to EDDN (dry: %t, test: %t)",
		upld.DryRun,
		upld.TestUrl)
	eddnTraceMsg(jump)
	err = upld.Send(eddn.Sjournal, jump)
	logEddn.Log(l.Ltrace, "done with journal event to EDDN")
	eddnSendErr(err, jump)
}

func eddnSendCommodities(upld *eddn.Upload, jStat map[string]interface{}) {
	stat := ggja.Obj{Bare: jStat, OnError: ggjaFailLogErr}
	tstr := stat.MStr("timestamp")
	if len(tstr) == 0 {
		return
	}
	logEddn.Logf(l.Ldebug, "send commodities to EDDN (dry: %t, test: %t)",
		upld.DryRun,
		upld.TestUrl)
	cmdt := eddn.NewMessage(tstr)
	err := eddn.SetCommoditiesJ(cmdt, jStat)
	if err != nil {
		log.Logf(l.Lerror, "eddn commodities: %s", err)
		return
	}
	eddnTraceMsg(cmdt)
	err = upld.Send(eddn.Scommodity, cmdt)
	logEddn.Log(l.Ltrace, "done with commodities to EDDN")
	eddnSendErr(err, cmdt)
}

func eddnSendShipyard(upld *eddn.Upload, jShy map[string]interface{}) {
	shy := ggja.Obj{Bare: jShy, OnError: ggjaFailLogErr}
	tstr := shy.MStr("timestamp")
	if len(tstr) == 0 {
		return
	}
	logEddn.Logf(l.Ldebug, "send shipyard to EDDN (dry: %t, test: %t)",
		upld.DryRun,
		upld.TestUrl)
	cmdt := eddn.NewMessage(tstr)
	err := eddn.SetShipyardJ(cmdt, jShy)
	if err != nil {
		log.Logf(l.Lerror, "eddn shipyard: %s", err)
		return
	}
	eddnTraceMsg(cmdt)
	err = upld.Send(eddn.Sshipyard, cmdt)
	logEddn.Log(l.Ltrace, "done with shipyard to EDDN")
	eddnSendErr(err, cmdt)
}
