package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"git.fractalqb.de/fractalqb/ggja"
	l "git.fractalqb.de/fractalqb/qblog"
	"github.com/CmdrVasquess/BCplus/cmdr"
	"github.com/CmdrVasquess/BCplus/galaxy"
	"github.com/CmdrVasquess/watched"
)

func takeTimeFromName(jfnm string) (time.Time, error) {
	jfnm = jfnm[8:20]
	res, err := time.ParseInLocation("060102150405", jfnm, time.Local)
	return res, err
}

func spoolJouranls(jdir string, startAfter time.Time) string {
	log.Logf(l.Ldebug, "spool journal events after %s", startAfter)
	rddir, err := os.Open(jdir)
	if err != nil {
		log.Log(l.Lerror, "fail to scan journal-dir: ", err)
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
	log.Logf(l.Linfo, "reading missed events from '%s'", jfnm)
	jf, err := os.Open(jfnm)
	if err != nil {
		log.Logf(l.Lerror, "cannot open journal: %s", err)
		return
	}
	defer jf.Close()
	scn := bufio.NewScanner(jf)
	for scn.Scan() {
		journalEvent(scn.Bytes())
	}
}

type jEventHdlr func(ts time.Time, evt ggja.GenObj) (backlog bool)

var jEventHdlrs = map[string]jEventHdlr{
	"Commander":        jevtCommander,
	"Docked":           jevtDocked,
	"Fileheader":       jevtFileheader,
	"FSDJump":          jevtFsdJump,
	"Location":         jevtLocation,
	"LoadGame":         jevtLoadGame,
	"Loadout":          jevtLoadout,
	"Materials":        jevtMaterials,
	"MissionAbandoned": jevtMissionAbandoned,
	"MissionAccepted":  jevtMissionAccepted,
	"MissionCompleted": jevtMissionCompleted,
	"Missions":         jevtMissions,
	"Progress":         jevtProgress,
	"Rank":             jevtRank,
	"Reputation":       jevtReputation,
	"Scan":             jevtScan,
	"ShipyardBuy":      jevtShipyardBuy,
	"ShipyardNew":      jevtShipyardNew,
	"ShipyardSell":     jevtShipyardSell,
	"ShipyardSwap":     jevtShipyardSwap,
	"ShipyardTransfer": jevtShipyardTransfer,
	"StartJump":        jevtStartJump,
	"SupercruiseExit":  jevtSupercruiseExit,
	"Undocked":         jevtUndocked,
}

var jevtBacklog []ggja.GenObj

func journalEvent(jLine []byte) {
	eJson := make(ggja.GenObj)
	err := json.Unmarshal(jLine, &eJson)
	if err != nil {
		log.Logf(l.Lerror, "cannot parse journal event: %s: %s", err, string(jLine))
		return
	}
	defer func() {
		if p := recover(); p != nil {
			log.Log(l.Lerror, "recover journal panic:", p)
		}
	}()
	evt := ggja.Obj{Bare: eJson}
	ets := evt.MTime("timestamp")
	enm := evt.Str("event", "")
	if len(enm) == 0 {
		log.Logf(l.Lerror, "no event name in journal event: %s", string(jLine))
		return
	}
	if ets.Before(bcpState.LastEDEvent) {
		log.Logf(l.Lwarn, "ignore historic event '%s' @%s <= %s",
			enm,
			ets.Format(time.RFC3339),
			bcpState.LastEDEvent.Format(time.RFC3339))
		switch enm {
		case "Commander":
			cmdr := evt.MStr("Name")
			switchToCommander(cmdr)
		case "LoadGame":
			cmdr := evt.MStr("Commander")
			switchToCommander(cmdr)
		case "Fileheader":
			switchToCommander("")
		}
		return
	}
	if theCmdr != nil {
		jEventMacro(enm, theCmdr.JStatFlags)
	}
	hdlr, ok := jEventHdlrs[enm]
	if ok {
		log.Logf(l.Ldebug, "dispatch to '%s' handler", enm)
		backlog := hdlr(ets, eJson)
		if backlog {
			log.Log(l.Ldebug, "putting event to backlog")
			jevtBacklog = append(jevtBacklog, eJson)
		} else {
			bcpState.LastEDEvent = ets
			if len(jevtBacklog) > 0 {
				var nbl []ggja.GenObj
				for _, jEvt := range jevtBacklog {
					evt := ggja.Obj{Bare: jEvt}
					ets := evt.MTime("timestamp")
					enm := evt.MStr("event")
					hdlr, _ = jEventHdlrs[enm]
					log.Logf(l.Ldebug, "dispatch from backlog to '%s' handler", enm)
					if hdlr(ets, jEvt) {
						log.Log(l.Ldebug, "keeping event in backlog")
						nbl = append(nbl, jEvt)
					}
				}
				jevtBacklog = nbl
			}
		}
	} else {
		log.Logf(l.Ldebug, "no handler for event '%s'", enm)
	}
}

func sysByName(name string, coos *ggja.Arr) (res *galaxy.System) {
	var err error
	if coos == nil {
		res, err = theGalaxy.MustSystem(name, nil)
	} else {
		res, err = theGalaxy.MustSystemCoos(name,
			coos.MF64(0), coos.MF64(1), coos.MF64(2),
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

func jevtFileheader(ts time.Time, evt ggja.GenObj) bool {
	stateLock.Lock()
	defer stateLock.Unlock()
	switchToCommander("")
	return false
}

func jevtCommander(ts time.Time, jEvt ggja.GenObj) bool {
	evt := ggja.Obj{Bare: jEvt}
	name := evt.MStr("Name")
	stateLock.Lock()
	defer stateLock.Unlock()
	switchToCommander(name)
	return false
}

func jevtDocked(ts time.Time, jEvt ggja.GenObj) bool {
	if theCmdr == nil {
		return true
	}
	stateLock.Lock()
	defer stateLock.Unlock()
	evt := ggja.Obj{Bare: jEvt}
	ssys := sysByName(evt.MStr("StarSystem"), nil)
	port, err := ssys.MustPart(galaxy.Port, evt.MStr("StationName"))
	theCmdr.Loc.SysId = ssys.Id
	theCmdr.Loc.LocId = port.Id
	theCmdr.Loc.Docked = true
	if err != nil {
		log.Panic(err)
	}
	if eddnMode != flagEddnOff {
		go eddnSendJournal(theEddn, ts, jEvt, ssys)
	}
	return false
}

func jevtUndocked(ts time.Time, evt ggja.GenObj) bool {
	if theCmdr == nil {
		return true
	}
	stateLock.Lock()
	defer stateLock.Unlock()
	// TODO compare station/port; problem unreliable system id
	theCmdr.Loc.Docked = false
	return false
}

func jevtLoadGame(ts time.Time, jEvt ggja.GenObj) bool {
	evt := ggja.Obj{Bare: jEvt}
	name := evt.MStr("Commander")
	stateLock.Lock()
	defer stateLock.Unlock()
	switchToCommander(name)
	theCmdr.Creds = evt.Int64("Credits", theCmdr.Creds)
	theCmdr.Loan = evt.Int64("Loan", theCmdr.Loan)
	ship := theCmdr.MustHaveShip(evt.MInt("ShipID"), evt.MStr("Ship"))
	theCmdr.InShip = ship.Id
	ship.BerthLoc = 0
	ship.Ident = evt.Str("ShipIdent", ship.Ident)
	ship.Name = evt.Str("ShipName", ship.Name)
	return false
}

func jevtLoadout(ts time.Time, jEvt ggja.GenObj) bool {
	if theCmdr == nil {
		return true
	}
	evt := ggja.Obj{Bare: jEvt}
	stateLock.Lock()
	defer stateLock.Unlock()
	ship := theCmdr.MustHaveShip(evt.MInt("ShipID"), evt.MStr("Ship"))
	theCmdr.InShip = ship.Id
	ship.BerthLoc = 0
	ship.Ident = evt.Str("ShipIdent", ship.Ident)
	ship.Name = evt.Str("ShipName", ship.Name)
	ship.Health = evt.F64("Health", ship.Health)
	ship.Rebuy = evt.Int("Rebuy", ship.Rebuy)
	ship.HullValue = evt.Int("HullValue", ship.HullValue)
	ship.ModuleValue = evt.Int("ModulesValue", ship.ModuleValue)
	return false
}

func jevtMaterials(ts time.Time, jEvt ggja.GenObj) bool {
	if theCmdr == nil {
		return true
	}
	stateLock.Lock()
	defer stateLock.Unlock()
	evt := ggja.Obj{Bare: jEvt}
	setMats := func(kind cmdr.MatKind, matls ggja.GenArr) {
		for _, tmp := range matls {
			eMat := ggja.Obj{Bare: tmp.(ggja.GenObj), OnError: evt.OnError}
			matKey := cmdr.MatDefine(kind, eMat.MStr("Name"))
			if cMat, ok := theCmdr.Mats[matKey]; ok {
				cMat.Have = eMat.MInt("Count")
			} else {
				cMat = &cmdr.MatState{Have: eMat.MInt("Count")}
				theCmdr.Mats[matKey] = cMat
			}
		}
	}
	setMats(cmdr.MatRaw, evt.Arr("Raw").Bare)
	setMats(cmdr.MatMan, evt.Arr("Manufactured").Bare)
	setMats(cmdr.MatEnc, evt.Arr("Encoded").Bare)
	return false
}

func jevtMissionAbandoned(ts time.Time, jEvt ggja.GenObj) bool {
	if theCmdr == nil {
		return true
	}
	stateLock.Lock()
	defer stateLock.Unlock()
	evt := ggja.Obj{Bare: jEvt}
	mid := evt.MUint32("MissionID")
	mission := theCmdr.Missions[mid]
	var impact float32
	if mission == nil {
		impact = 1.5
	} else {
		delete(theCmdr.Missions, mid)
		impact = mission.Reputation
	}
	level := theCmdr.MinorRep[mission.Faction]
	theCmdr.MinorRep[mission.Faction] = level - impact
	return false
}

func jevtMissionAccepted(ts time.Time, jEvt ggja.GenObj) bool {
	if theCmdr == nil {
		return true
	}
	stateLock.Lock()
	defer stateLock.Unlock()
	evt := ggja.Obj{Bare: jEvt}
	mid := evt.MUint32("MissionID")
	repStr := evt.MStr("Reputation")
	var rep float32
	switch repStr {
	case "None":
		rep = 0
	case "Low":
		rep = 1
	case "Med":
		rep = 2
	case "High":
		rep = 3
	}
	theCmdr.Missions[mid] = &cmdr.Mission{
		Faction:    evt.MStr("Faction"),
		Reputation: rep,
	}
	return false
}

func jevtMissionCompleted(ts time.Time, jEvt ggja.GenObj) bool {
	if theCmdr == nil {
		return true
	}
	stateLock.Lock()
	defer stateLock.Unlock()
	evt := ggja.Obj{Bare: jEvt}
	mid := evt.MUint32("MissionID")
	mission := theCmdr.Missions[mid]
	var impact float32
	if mission == nil {
		impact = 1.5
	} else {
		delete(theCmdr.Missions, mid)
		impact = mission.Reputation
	}
	if effs := evt.Arr("FactionEffects"); effs != nil {
		for i := range effs.Bare {
			eff := effs.MObj(i)
			faction := eff.MStr("Faction")
			if len(faction) == 0 {
				continue
			}
			rep := eff.Str("Reputation", "")
			if rep == "" {
				continue
			}
			switch rep {
			case "UpGood":
				level := theCmdr.MinorRep[faction]
				theCmdr.MinorRep[faction] = level + impact
			case "DownBad":
				level := theCmdr.MinorRep[faction]
				theCmdr.MinorRep[faction] = level - impact
			default:
				log.Logf(l.Lwarn, "illegal mission reputation trend: '%s'", rep)
			}
		}
	}
	return false
}

func jevtMissions(ts time.Time, jEvt ggja.GenObj) bool {
	if theCmdr == nil {
		return true
	}
	stateLock.Lock()
	defer stateLock.Unlock()
	evt := ggja.Obj{Bare: jEvt}
	var actvIds, rmIds []uint32
	actv := evt.MArr("Active")
	for i := range actv.Bare {
		actvIds = append(actvIds, actv.MUint32(i))
	}
NEXT_CHECK_RM:
	for k := range theCmdr.Missions {
		for _, a := range actvIds {
			if k == a {
				continue NEXT_CHECK_RM
			}
		}
		rmIds = append(rmIds, k)
	}
	for _, rm := range rmIds {
		delete(theCmdr.Missions, rm)
	}
	return false
}

func jevtRank(ts time.Time, jEvt ggja.GenObj) bool {
	if theCmdr == nil {
		return true
	}
	stateLock.Lock()
	defer stateLock.Unlock()
	evt := ggja.Obj{Bare: jEvt}
	theCmdr.Ranks.Imps.Level = evt.Int("Empire", theCmdr.Ranks.Imps.Level)
	theCmdr.Ranks.Feds.Level = evt.Int("Federation", theCmdr.Ranks.Feds.Level)
	theCmdr.Ranks.Combat.Level = evt.Int("Combat", theCmdr.Ranks.Combat.Level)
	theCmdr.Ranks.Trade.Level = evt.Int("Trade", theCmdr.Ranks.Trade.Level)
	theCmdr.Ranks.Explore.Level = evt.Int("Explore", theCmdr.Ranks.Explore.Level)
	theCmdr.Ranks.CQC.Level = evt.Int("CQC", theCmdr.Ranks.CQC.Level)
	return false
}

func jevtReputation(ts time.Time, jEvt ggja.GenObj) bool {
	if theCmdr == nil {
		return true
	}
	stateLock.Lock()
	defer stateLock.Unlock()
	evt := ggja.Obj{Bare: jEvt}
	theCmdr.Rep.Imps = evt.F32("Empire", theCmdr.Rep.Imps)
	theCmdr.Rep.Feds = evt.F32("Federation", theCmdr.Rep.Feds)
	theCmdr.Rep.Allis = evt.F32("Alliance", theCmdr.Rep.Allis)
	return false
}

func jevtProgress(ts time.Time, jEvt ggja.GenObj) bool {
	if theCmdr == nil {
		return true
	}
	stateLock.Lock()
	defer stateLock.Unlock()
	evt := ggja.Obj{Bare: jEvt}
	theCmdr.Ranks.Imps.Progress = evt.Int("Empire", theCmdr.Ranks.Imps.Progress)
	theCmdr.Ranks.Feds.Progress = evt.Int("Federation", theCmdr.Ranks.Feds.Progress)
	theCmdr.Ranks.Combat.Progress = evt.Int("Combat", theCmdr.Ranks.Combat.Progress)
	theCmdr.Ranks.Trade.Progress = evt.Int("Trade", theCmdr.Ranks.Trade.Progress)
	theCmdr.Ranks.Explore.Progress = evt.Int("Explore", theCmdr.Ranks.Explore.Progress)
	theCmdr.Ranks.CQC.Progress = evt.Int("CQC", theCmdr.Ranks.CQC.Progress)
	return false
}

func jevtLocation(ts time.Time, jEvt ggja.GenObj) bool {
	if theCmdr == nil {
		return true
	}
	evt := ggja.Obj{Bare: jEvt}
	sysNm := evt.MStr("StarSystem")
	coos := evt.MArr("StarPos")
	station := evt.Str("StationName", "")
	stateLock.Lock()
	defer stateLock.Unlock()
	xa := theGalaxy.XaBegin()
	defer xa.Rollback()
	ssys := sysByName(sysNm, coos)
	theCmdr.Loc.SysId = ssys.Id
	theCmdr.Loc.Docked = evt.Bool("Docked", false)
	if len(station) > 0 {
		stn := partFromSys(ssys, galaxy.Port, station)
		putStn := false
		theCmdr.Loc.LocId = stn.Id
		if evt.Str("StationType", "") == "SurfaceStation" {
			if stn.Height != 0 {
				stn.Height = 0
				putStn = true
			}
		} else if stn.Height == 0 {
			stn.Height = galaxy.NaN32
			putStn = true
		}
		if evt.Str("BodyType", "") == "Planet" {
			pnm := evt.MStr("Body")
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
	} else if bdynm := evt.Str("Body", ""); len(bdynm) > 0 {
		btype := evt.MStr("BodyType")
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
		go eddnSendJournal(theEddn, ts, jEvt, ssys)
	}
	return false
}

func jevtFsdJump(ts time.Time, jEvt ggja.GenObj) bool {
	if theCmdr == nil {
		return true
	}
	stateLock.Lock()
	defer stateLock.Unlock()
	evt := ggja.Obj{Bare: jEvt}
	sysNm := evt.MStr("StarSystem")
	coos := evt.MArr("StarPos")
	ssys := sysByName(sysNm, coos)
	theCmdr.Loc.SysId = ssys.Id
	theCmdr.Loc.LocId = 0
	theCmdr.Loc.Docked = false
	theCmdr.Jumps.Add(ssys.Id, ts)
	if eddnMode != flagEddnOff {
		go eddnSendJournal(theEddn, ts, jEvt, ssys)
	}
	return false
}

func jevtScan(ts time.Time, jEvt ggja.GenObj) bool {
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
			go eddnSendJournal(theEddn, ts, jEvt, ssys)
		}
	}
	return false
}

func jevtShipyardBuy(ts time.Time, jEvt ggja.GenObj) bool {
	if theCmdr == nil {
		return true
	}
	evt := ggja.Obj{Bare: jEvt}
	storeId := evt.MInt("StoreShipID")
	if storeId != theCmdr.InShip {
		log.Logf(l.Lwarn, "inconsistent ids for current ship bc+:%d / event:%d",
			theCmdr.InShip, storeId)
	}
	stateLock.Lock()
	defer stateLock.Unlock()
	ship := theCmdr.MustHaveShip(storeId, evt.MStr("ShipType"))
	ship.BerthLoc = theCmdr.Loc.LocId
	theCmdr.InShip = cmdr.NoShip
	return false
}

func jevtShipyardNew(ts time.Time, jEvt ggja.GenObj) bool {
	if theCmdr == nil {
		return true
	}
	evt := ggja.Obj{Bare: jEvt}
	ship := &cmdr.Ship{
		Id:     evt.MInt("NewShipID"),
		Type:   evt.MStr("ShipType"),
		Bought: ts,
	}
	stateLock.Lock()
	defer stateLock.Unlock()
	if theCmdr.InShip > cmdr.NoShip {
		log.Log(l.Lwarn, "commander in a ship %d when new ship arrives",
			theCmdr.InShip)
		if s2s, ok := theCmdr.Ships[theCmdr.InShip]; ok {
			s2s.BerthLoc = theCmdr.Loc.LocId
		}
	}
	theCmdr.Ships[ship.Id] = ship
	theCmdr.InShip = ship.Id
	return false
}

func saveSoldShip(cmdr string, ship *cmdr.Ship) {
	var fnm string
	if len(ship.Ident) > 0 {
		if len(ship.Name) > 0 {
			fnm = fmt.Sprintf("%s - %s", ship.Name, ship.Ident)
		} else {
			fnm = ship.Ident
		}
	} else if len(ship.Name) > 0 {
		fnm = ship.Name
	} else {
		fnm = ship.Type
	}
	fnm = fmt.Sprintf("sold_ships/Ship%d %s.json", ship.Id, fnm)
	fnm = cmdrFile(cmdr, fnm)
	f, err := os.Create(fnm)
	if err != nil {
		log.Logf(l.Lwarn, "cannot save sold ship %d %s", ship.Id, fnm)
		return
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	err = enc.Encode(ship)
}

func jevtShipyardSell(ts time.Time, jEvt ggja.GenObj) bool {
	if theCmdr == nil {
		return true
	}
	evt := ggja.Obj{Bare: jEvt}
	sid := evt.MInt("SellShipID")
	stateLock.Lock()
	defer stateLock.Unlock()
	if theCmdr.InShip == sid {
		log.Logf(l.Lwarn, "commander in sold ship %d", sid)
		theCmdr.InShip = cmdr.NoShip
	}
	if ship, ok := theCmdr.Ships[sid]; ok && ship != nil {
		go saveSoldShip(theCmdr.Name, ship)
	}
	delete(theCmdr.Ships, sid)
	return false
}

func jevtShipyardSwap(ts time.Time, jEvt ggja.GenObj) bool {
	if theCmdr == nil {
		return true
	}
	evt := ggja.Obj{Bare: jEvt}
	stateLock.Lock()
	defer stateLock.Unlock()
	ship := theCmdr.MustHaveShip(
		evt.MInt("StoreShipID"),
		evt.MStr("StoreOldShip"),
	)
	ship.BerthLoc = theCmdr.Loc.LocId
	ship = theCmdr.MustHaveShip(
		evt.MInt("ShipID"),
		evt.MStr("ShipType"),
	)
	ship.BerthLoc = 0
	theCmdr.InShip = ship.Id
	return false
}

func jevtShipyardTransfer(ts time.Time, jEvt ggja.GenObj) bool {
	if theCmdr == nil {
		return true
	}
	evt := ggja.Obj{Bare: jEvt}
	sid := evt.MInt("ShipID")
	stateLock.Lock()
	defer stateLock.Unlock()
	if theCmdr.InShip == sid {
		log.Logf(l.Lwarn, "commander in transferred ship %d", sid)
		theCmdr.InShip = cmdr.NoShip
	}
	ship := theCmdr.MustHaveShip(sid, evt.MStr("ShipType"))
	ship.BerthLoc = theCmdr.Loc.LocId
	return false
}

func jevtStartJump(ts time.Time, jEvt ggja.GenObj) bool {
	if theCmdr == nil {
		return true
	}
	evt := ggja.Obj{Bare: jEvt}
	jty := evt.MStr("JumpType")
	if jty == "Supercruise" {
		stateLock.Lock()
		defer stateLock.Unlock()
		theCmdr.Loc.LocId = 0
		theCmdr.Loc.Docked = false
	}
	return false
}

func jevtSupercruiseExit(ts time.Time, jEvt ggja.GenObj) bool {
	if theCmdr == nil {
		return true
	}
	// TODO NYI
	return false
}
