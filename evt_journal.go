package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/CmdrVasquess/BCplus/galaxy"
	"github.com/CmdrVasquess/watched"
	l "git.fractalqb.de/fractalqb/qblog"
)

func takeTimeFromName(jfnm string) (time.Time, error) {
	jfnm = jfnm[8:20]
	res, err := time.ParseInLocation("060102150405", jfnm, time.Local)
	return res, err
}

func spoolJouranls(jdir string, startAfter time.Time) string {
	log.Logf(l.Debug, "spool journal events after %s", startAfter)
	rddir, err := os.Open(jdir)
	if err != nil {
		log.Log(l.Error, "fail to scan journal-dir: ", err)
		return ""
	}
	defer rddir.Close()
	var jfls []string
	var maxBeforeStart time.Time // Zero should be a reasonable start
	var jBefore string
	infos, err := rddir.Readdir(1)
	for len(infos) > 0 && err == nil {
		info := infos[0]
		if watched.IsJournalFile(info.Name()) {
			jft, err := takeTimeFromName(info.Name())
			if err != nil {
				continue
			}
			if jft.After(startAfter) {
				jfls = append(jfls, info.Name())
			} else if jft.After(maxBeforeStart) {
				maxBeforeStart = info.ModTime()
				jBefore = info.Name()
			}
		}
		infos, err = rddir.Readdir(1)
	}
	if len(jBefore) > 0 {
		jfls = append(jfls, jBefore)
	}
	if len(jfls) == 0 {
		return ""
	}
	sort.Strings(jfls)
	oldEDDN := eddnMode
	eddnMode = flagEddnOff
	defer func() { eddnMode = oldEDDN }()
	for _, j := range jfls[:len(jfls)-1] {
		jfnm := filepath.Join(jdir, j)
		readJournal(jfnm)
	}
	return jfls[len(jfls)-1]
}

func readJournal(jfnm string) {
	log.Logf(l.Info, "reading missed events from '%s'", jfnm)
	jf, err := os.Open(jfnm)
	if err != nil {
		log.Logf(l.Error, "cannot open journal: %s", err)
		return
	}
	defer jf.Close()
	scn := bufio.NewScanner(jf)
	for scn.Scan() {
		journalEvent(scn.Bytes())
	}
}

type jEvent = map[string]interface{}

type jEventHdlr func(ts time.Time, evt jEvent) (backlog bool)

var jEventHdlrs = map[string]jEventHdlr{
	"Commander":       jevtCommander,
	"Docked":          jevtDocked,
	"Fileheader":      jevtFileheader,
	"FSDJump":         jevtFsdJump,
	"Location":        jevtLocation,
	"LoadGame":        jevtLoadGame,
	"Progress":        jevtProgress,
	"Rank":            jevtRank,
	"Reputation":      jevtReputation,
	"Scan":            jevtScan,
	"StartJump":       jevtStartJump,
	"SupercruiseExit": jevtSupercruiseExit,
	"Undocked":        jevtUndocked,
}

var jevtBacklog []jEvent

func journalEvent(jLine []byte) {
	evt := make(jEvent)
	err := json.Unmarshal(jLine, &evt)
	if err != nil {
		log.Logf(l.Error, "cannot parse journal event: %s: %s", err, string(jLine))
		return
	}
	defer func() {
		if p := recover(); p != nil {
			log.Logf(l.Error, "recover journal panic:", p)
		}
	}()
	ets := jrmgTs(evt, "timestamp")
	enm := jrgStr(evt, "event", "")
	if len(enm) == 0 {
		log.Logf(l.Error, "no event name in journal event: %s", string(jLine))
		return
	}
	if ets.Before(bcpState.LastEDEvent) {
		log.Logf(l.Warn, "ignore historic event '%s' @%s <= %s",
			enm,
			ets.Format(time.RFC3339),
			bcpState.LastEDEvent.Format(time.RFC3339))
		switch enm {
		case "Commander":
			cmdr := jrmgStr(evt, "Name")
			switchToCommander(cmdr)
		case "LoadGame":
			cmdr := jrmgStr(evt, "Commander")
			switchToCommander(cmdr)
		case "Fileheader":
			switchToCommander("")
		}
		return
	}
	hdlr, ok := jEventHdlrs[enm]
	if ok {
		log.Logf(l.Debug, "dispatch to '%s' handler", enm)
		backlog := hdlr(ets, evt)
		if backlog {
			log.Log(l.Debug, "putting event to backlog")
			jevtBacklog = append(jevtBacklog, evt)
		} else {
			bcpState.LastEDEvent = ets
			if len(jevtBacklog) > 0 {
				var nbl []jEvent
				for _, evt := range jevtBacklog {
					ets := jrmgTs(evt, "timestamp")
					enm := jrgStr(evt, "event", "")
					hdlr, _ = jEventHdlrs[enm]
					log.Logf(l.Debug, "dispatch from backlog to '%s' handler", enm)
					if hdlr(ets, evt) {
						log.Log(l.Debug, "keeping event in backlog")
						nbl = append(nbl, evt)
					}
				}
				jevtBacklog = nbl
			}
		}
	} else {
		log.Logf(l.Debug, "no handler for event '%s'", enm)
	}
}

func sysByName(name string, coos interface{}) (res *galaxy.System) {
	var err error
	if coos == nil {
		res, err = theGalaxy.MustSystem(name, nil)
	} else {
		tmp := coos.([]interface{})
		res, err = theGalaxy.MustSystemCoos(name,
			tmp[galaxy.Xk].(float64), tmp[galaxy.Yk].(float64), tmp[galaxy.Zk].(float64),
			nil)
	}
	if err != nil {
		log.Panic(err)
	}
	return res
}

func partFromSys(sys *galaxy.System, typ galaxy.PartType, partName string) *galaxy.SysPart {
	res, err := sys.MustPart(typ, partName)
	if err != nil {
		log.Panic(err)
	}
	return res
}

func jevtFileheader(ts time.Time, evt jEvent) bool {
	stateLock.Lock()
	defer stateLock.Unlock()
	switchToCommander("")
	return false
}

func jevtCommander(ts time.Time, evt jEvent) bool {
	name := jrmgStr(evt, "Name")
	stateLock.Lock()
	defer stateLock.Unlock()
	switchToCommander(name)
	return false
}

func jevtDocked(ts time.Time, evt jEvent) bool {
	if theCmdr == nil {
		return true
	}
	stateLock.Lock()
	defer stateLock.Unlock()
	ssys := sysByName(jrmgStr(evt, "StarSystem"), nil)
	port, err := ssys.MustPart(galaxy.Port, jrmgStr(evt, "StationName"))
	theCmdr.Loc.SysId = ssys.Id
	theCmdr.Loc.LocId = port.Id
	theCmdr.Loc.Docked = true
	if err != nil {
		log.Panic(err)
	}
	if eddnMode != flagEddnOff {
		go eddnSendJournal(theEddn, ts, evt, ssys)
	}
	return false
}

func jevtUndocked(ts time.Time, evt jEvent) bool {
	if theCmdr == nil {
		return true
	}
	stateLock.Lock()
	defer stateLock.Unlock()
	// TODO compare station/port; problem unreliable system id
	theCmdr.Loc.Docked = false
	return false
}

func jevtLoadGame(ts time.Time, evt jEvent) bool {
	name := jrmgStr(evt, "Commander")
	stateLock.Lock()
	defer stateLock.Unlock()
	switchToCommander(name)
	jruInt64(&theCmdr.Creds, evt, "Credits")
	jruInt64(&theCmdr.Loan, evt, "Loan")
	return false
}

func jevtRank(ts time.Time, evt jEvent) bool {
	if theCmdr == nil {
		return true
	}
	stateLock.Lock()
	defer stateLock.Unlock()
	jruInt(&theCmdr.Ranks.Imps.Level, evt, "Empire")
	jruInt(&theCmdr.Ranks.Feds.Level, evt, "Federation")
	jruInt(&theCmdr.Ranks.Combat.Level, evt, "Combat")
	jruInt(&theCmdr.Ranks.Trade.Level, evt, "Trade")
	jruInt(&theCmdr.Ranks.Explore.Level, evt, "Explore")
	jruInt(&theCmdr.Ranks.CQC.Level, evt, "CQC")
	return false
}

func jevtReputation(ts time.Time, evt jEvent) bool {
	if theCmdr == nil {
		return true
	}
	stateLock.Lock()
	defer stateLock.Unlock()
	jruF32(&theCmdr.Rep.Imps, evt, "Empire")
	jruF32(&theCmdr.Rep.Feds, evt, "Federation")
	jruF32(&theCmdr.Rep.Allis, evt, "Alliance")
	return false
}

func jevtProgress(ts time.Time, evt jEvent) bool {
	if theCmdr == nil {
		return true
	}
	stateLock.Lock()
	defer stateLock.Unlock()
	jruInt(&theCmdr.Ranks.Imps.Progress, evt, "Empire")
	jruInt(&theCmdr.Ranks.Feds.Progress, evt, "Federation")
	jruInt(&theCmdr.Ranks.Combat.Progress, evt, "Combat")
	jruInt(&theCmdr.Ranks.Trade.Progress, evt, "Trade")
	jruInt(&theCmdr.Ranks.Explore.Progress, evt, "Explore")
	jruInt(&theCmdr.Ranks.CQC.Progress, evt, "CQC")
	return false
}

func jevtLocation(ts time.Time, evt jEvent) bool {
	if theCmdr == nil {
		return true
	}
	sysNm := jrmgStr(evt, "StarSystem")
	coos, ok := evt["StarPos"]
	if !ok {
		panic(fmt.Errorf("system '%s' without coordinates", sysNm))
	}
	station := jrgStr(evt, "StationName", "")
	stateLock.Lock()
	defer stateLock.Unlock()
	xa := theGalaxy.XaBegin()
	defer xa.Rollback()
	ssys := sysByName(sysNm, coos)
	theCmdr.Loc.SysId = ssys.Id
	theCmdr.Loc.Docked = jrgBool(evt, "Docked", false)
	if len(station) > 0 {
		stn := partFromSys(ssys, galaxy.Port, station)
		putStn := false
		theCmdr.Loc.LocId = stn.Id
		if jrgStr(evt, "StationType", "") == "SurfaceStation" {
			if stn.Height != 0 {
				stn.Height = 0
				putStn = true
			}
		} else if stn.Height == 0 {
			stn.Height = galaxy.NaN32
			putStn = true
		}
		if jrgStr(evt, "BodyType", "") == "Planet" {
			pnm := jrmgStr(evt, "Body")
			pnm = ssys.LocalName(pnm)
			planet := partFromSys(ssys, galaxy.Planet, pnm)
			if stn.TiedTo != planet.Id {
				stn.TiedTo = planet.Id
				putStn = true
			}
		}
		if putStn {
			if _, err := theGalaxy.PutSysPart(stn); err != nil {
				log.Panic(err)
			}
		}
	} else if bdynm := jrgStr(evt, "Body", ""); len(bdynm) > 0 {
		btype := jrmgStr(evt, "BodyType")
		var body *galaxy.SysPart
		switch btype {
		case "Planet":
			body = partFromSys(ssys, galaxy.Planet, ssys.LocalName(bdynm))
		case "Star":
			body = partFromSys(ssys, galaxy.Star, ssys.LocalName(bdynm))
		}
		if body != nil {
			theCmdr.Loc.LocId = body.Id
		}
	} else {
		theCmdr.Loc.LocId = 0
	}
	xa.Commit()
	if eddnMode != flagEddnOff {
		go eddnSendJournal(theEddn, ts, evt, ssys)
	}
	return false
}

func jevtFsdJump(ts time.Time, evt jEvent) bool {
	if theCmdr == nil {
		return true
	}
	stateLock.Lock()
	defer stateLock.Unlock()
	sysNm := jrmgStr(evt, "StarSystem")
	coos, ok := evt["StarPos"]
	if !ok {
		panic(fmt.Errorf("system '%s' without coordinates", sysNm))
	}
	ssys := sysByName(sysNm, coos)
	theCmdr.Loc.SysId = ssys.Id
	theCmdr.Loc.LocId = 0
	theCmdr.Loc.Docked = false
	theCmdr.Jumps.Add(ssys.Id, ts)
	if eddnMode != flagEddnOff {
		go eddnSendJournal(theEddn, ts, evt, ssys)
	}
	return false
}

func jevtScan(ts time.Time, evt jEvent) bool {
	if theCmdr == nil {
		return true
	}
	// TODO update cmdr location
	if theCmdr.Loc.SysId > 0 {
		ssys, err := theGalaxy.GetSystem(theCmdr.Loc.SysId, nil)
		if err != nil {
			panic(err)
		}
		if ssys == nil {
			return false
		}
		if eddnMode != flagEddnOff {
			go eddnSendJournal(theEddn, ts, evt, ssys)
		}
	}
	return false
}

func jevtStartJump(ts time.Time, evt jEvent) bool {
	if theCmdr == nil {
		return true
	}
	jty := jrmgStr(evt, "JumpType")
	if jty == "Supercruise" {
		stateLock.Lock()
		defer stateLock.Unlock()
		theCmdr.Loc.LocId = 0
		theCmdr.Loc.Docked = false
	}
	return false
}

func jevtSupercruiseExit(ts time.Time, evt jEvent) bool {
	// TODO NYI
	return false
}
