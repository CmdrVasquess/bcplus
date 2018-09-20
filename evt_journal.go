package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/CmdrVasquess/BCplus/webui"

	"git.fractalqb.de/fractalqb/ggja"
	l "git.fractalqb.de/fractalqb/qblog"
	"github.com/CmdrVasquess/BCplus/cmdr"
	"github.com/CmdrVasquess/BCplus/common"
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
	jevtSpooling = true
	defer func() {
		jevtSpooling = false
	}()
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

type jePost = uint32

const (
	jePostReload jePost = (1 << iota)
	jePostHdr
	jePostSysPop
)

type jeActn struct {
	hdlr    func(ts time.Time, evt ggja.Obj) jePost
	backlog bool
}

var jEvtMatCat = map[string]cmdr.MatKind{
	"Raw":          cmdr.MatRaw,
	"Manufactured": cmdr.MatMan,
	"Encoded":      cmdr.MatEnc,
}

var jEventHdl = map[string]jeActn{
	"Commander":         jeActn{jevtCommander, false},
	"Docked":            jeActn{jevtDocked, true},
	"Fileheader":        jeActn{jevtFileheader, false},
	"FSDJump":           jeActn{jevtFsdJump, true},
	"Location":          jeActn{jevtLocation, true},
	"LoadGame":          jeActn{jevtLoadGame, false},
	"Loadout":           jeActn{jevtLoadout, true},
	"MaterialCollected": jeActn{jevtMaterialCollected, true},
	"Materials":         jeActn{jevtMaterials, true},
	"MissionAbandoned":  jeActn{jevtMissionAbandoned, true},
	"MissionAccepted":   jeActn{jevtMissionAccepted, true},
	"MissionCompleted":  jeActn{jevtMissionCompleted, true},
	"Missions":          jeActn{jevtMissions, true},
	"Progress":          jeActn{jevtProgress, true},
	"Rank":              jeActn{jevtRank, true},
	"Reputation":        jeActn{jevtReputation, true},
	"Scan":              jeActn{jevtScan, true},
	"ShipyardBuy":       jeActn{jevtShipyardBuy, true},
	"ShipyardNew":       jeActn{jevtShipyardNew, true},
	"ShipyardSell":      jeActn{jevtShipyardSell, true},
	"ShipyardSwap":      jeActn{jevtShipyardSwap, true},
	"ShipTargeted":      jeActn{jevtShipTargeted, false},
	"ShipyardTransfer":  jeActn{jevtShipyardTransfer, true},
	"Shutdown":          jeActn{jevtShutdown, false},
	"StartJump":         jeActn{jevtStartJump, true},
	"SupercruiseExit":   jeActn{jevtSupercruiseExit, true},
	"Undocked":          jeActn{jevtUndocked, true},
}

var jevtSpooling = false
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
		log.Tracef("ignore historic event '%s' @%s <= %s",
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
	acnt, ok := jEventHdl[enm]
	if ok {
		if theCmdr == nil && acnt.backlog {
			log.Log(l.Ldebug, "putting event '%s' to backlog", enm)
			jevtBacklog = append(jevtBacklog, eJson)
		} else {
			if jevtSpooling {
				log.Debugf("spooling to '%s' handler", enm)
			} else {
				log.Debugf("dispatch to '%s' handler", enm)
			}
			var post jePost
			if len(jevtBacklog) > 0 {
				var nbl []ggja.GenObj
				for _, jEvt := range jevtBacklog {
					evt := ggja.Obj{Bare: jEvt}
					ets := evt.MTime("timestamp")
					enm := evt.MStr("event")
					blActn, _ := jEventHdl[enm]
					log.Logf(l.Ldebug, "dispatch from backlog to '%s' handler", enm)
					post |= blActn.hdlr(ets, ggja.Obj{Bare: jEvt})
				}
				jevtBacklog = nbl
			}
			post |= acnt.hdlr(ets, evt)
			bcpState.LastEDEvent = ets
			if !jevtSpooling {
				var cmd interface{}
				switch {
				case post&jePostReload == jePostReload:
					cmd = &webui.WsCmdLoad{
						WsCommand: webui.WsCommand{Cmd: webui.WsLoadCmd},
					}
				case post&jePostHdr == jePostHdr:
					cmd = webui.NewWsCmdUpdate(nil)
				}
				if cmd != nil {
					toWsClient <- cmd
				}
			}
		}
	} else {
		log.Debugf("no handler for event '%s'", enm)
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

func jevtFileheader(ts time.Time, evt ggja.Obj) jePost {
	stateLock.Lock()
	defer stateLock.Unlock()
	lang := evt.MStr("language")
	locale, dom := nameMaps.Lang.Map(nameMaps.LangEd, lang, nameMaps.LangLoc)
	nameMaps.Save()
	if dom >= 0 {
		nameMaps.Load(resDir, flagDDir, locale)
	} else {
		nameMaps.Load(resDir, flagDDir, common.DefaultLang)
	}
	switchToCommander("")
	return jePostReload
}

func jevtCommander(ts time.Time, evt ggja.Obj) jePost {
	name := evt.MStr("Name")
	stateLock.Lock()
	defer stateLock.Unlock()
	switchToCommander(name)
	return jePostReload
}

func jevtDocked(ts time.Time, evt ggja.Obj) jePost {
	stateLock.Lock()
	defer stateLock.Unlock()
	ssys := sysByName(evt.MStr("StarSystem"), nil)
	port, err := ssys.MustPart(galaxy.Port, evt.MStr("StationName"))
	theCmdr.Loc.SysId = ssys.Id
	theCmdr.Loc.LocId = port.Id
	theCmdr.Loc.Docked = true
	if err != nil {
		log.Panic(err)
	}
	eddnSendJournal(theEddn, ts, evt, ssys)
	return jePostHdr
}

func jevtUndocked(ts time.Time, evt ggja.Obj) jePost {
	stateLock.Lock()
	defer stateLock.Unlock()
	// TODO compare station/port; problem unreliable system id
	theCmdr.Loc.Docked = false
	return 0
}

func jevtLoadGame(ts time.Time, evt ggja.Obj) jePost {
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
	nameMaps.ShipType.PickL10n(evt, "Ship", "Ship_Localised")
	return jePostReload
}

func jevtLoadout(ts time.Time, evt ggja.Obj) jePost {
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
	return jePostHdr
}

func akkuMats(cat string, mat string, d int) {
	catId, ok := jEvtMatCat[cat]
	if !ok {
		log.Warnf("unknown material category: '%s'", cat)
		return
	}
	cmdrMat := cmdr.DefMaterial(catId, mat)
	if stock, ok := theCmdr.Mats[cmdrMat]; ok {
		stock.Have += d
	} else {
		stock := &cmdr.MatState{Have: d}
		theCmdr.Mats[cmdrMat] = stock
	}
}

func jevtMaterialCollected(ts time.Time, evt ggja.Obj) jePost {
	stateLock.Lock()
	defer stateLock.Unlock()
	cat := evt.MStr("Category")
	mat := evt.MStr("Name")
	bcpState.MatCats[mat] = cat
	loc := evt.Str("Name_Localised", "")
	if len(loc) > 0 {
		nameMaps.Material.SetL10n(mat, loc)
	}
	num := evt.MInt("Count")
	akkuMats(cat, mat, num)
	return 0
}

func jevtMaterials(ts time.Time, evt ggja.Obj) jePost {
	stateLock.Lock()
	defer stateLock.Unlock()
	setMats := func(cat string, kind cmdr.MatKind, matls ggja.GenArr) {
		for _, tmp := range matls {
			eMat := ggja.Obj{Bare: tmp.(ggja.GenObj), OnError: evt.OnError}
			matNm := eMat.MStr("Name")
			matKey := cmdr.DefMaterial(kind, matNm)
			if cMat, ok := theCmdr.Mats[matKey]; ok {
				cMat.Have = eMat.MInt("Count")
			} else {
				cMat = &cmdr.MatState{Have: eMat.MInt("Count")}
				theCmdr.Mats[matKey] = cMat
			}
			bcpState.MatCats[matNm] = cat
		}
	}
	setMats("Raw", cmdr.MatRaw, evt.Arr("Raw").Bare)
	setMats("Manufactured", cmdr.MatMan, evt.Arr("Manufactured").Bare)
	setMats("Encoded", cmdr.MatEnc, evt.Arr("Encoded").Bare)
	return 0
}

func jevtMissionAbandoned(ts time.Time, evt ggja.Obj) jePost {
	stateLock.Lock()
	defer stateLock.Unlock()
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
	return jePostSysPop
}

func splitMultiDestination(multiDest string) (dests []string) {
	tmp := strings.Split(multiDest, "$MISSIONUTIL_MULTIPLE_FINAL_SEPARATOR;")
	if len(tmp) < 2 {
		return tmp
	}
	dests = strings.Split(tmp[0], "$MISSIONUTIL_MULTIPLE_INNER_SEPARATOR;")
	dests = append(dests, tmp[1])
	return dests
}

func jevtMissionAccepted(ts time.Time, evt ggja.Obj) jePost {
	stateLock.Lock()
	defer stateLock.Unlock()
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
	mssn := &cmdr.Mission{
		Faction:    evt.MStr("Faction"),
		Title:      evt.MStr("LocalisedName"),
		Reputation: rep,
	}
	theCmdr.Missions[mid] = mssn
	if evt.MStr("Name") == "Mission_Sightseeing" {
		dests := splitMultiDestination(evt.MStr("DestinationSystem"))
		var (
			sys *galaxy.System
			err error
		)
		for _, d := range dests {
			sys, err = theGalaxy.MustSystem(d, sys)
			if err != nil {
				log.Panic(err)
			}
			mssn.Dests = append(mssn.Dests, sys.Id)
			if !galaxy.V3dValid(sys.Coos) {
				log.Warnf("system %d '%s' without coos", sys.Id, sys.Name)
			}
		}
	}
	return jePostSysPop
}

func jevtMissionCompleted(ts time.Time, evt ggja.Obj) jePost {
	stateLock.Lock()
	defer stateLock.Unlock()
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
	return jePostSysPop
}

func jevtMissions(ts time.Time, evt ggja.Obj) jePost {
	stateLock.Lock()
	defer stateLock.Unlock()
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
	return 0
}

func jevtRank(ts time.Time, evt ggja.Obj) jePost {
	stateLock.Lock()
	defer stateLock.Unlock()
	theCmdr.Ranks.Imps.Level = evt.Int("Empire", theCmdr.Ranks.Imps.Level)
	theCmdr.Ranks.Feds.Level = evt.Int("Federation", theCmdr.Ranks.Feds.Level)
	theCmdr.Ranks.Combat.Level = evt.Int("Combat", theCmdr.Ranks.Combat.Level)
	theCmdr.Ranks.Trade.Level = evt.Int("Trade", theCmdr.Ranks.Trade.Level)
	theCmdr.Ranks.Explore.Level = evt.Int("Explore", theCmdr.Ranks.Explore.Level)
	theCmdr.Ranks.CQC.Level = evt.Int("CQC", theCmdr.Ranks.CQC.Level)
	return 0
}

func jevtReputation(ts time.Time, evt ggja.Obj) jePost {
	stateLock.Lock()
	defer stateLock.Unlock()
	theCmdr.Rep.Imps = evt.F32("Empire", theCmdr.Rep.Imps)
	theCmdr.Rep.Feds = evt.F32("Federation", theCmdr.Rep.Feds)
	theCmdr.Rep.Allis = evt.F32("Alliance", theCmdr.Rep.Allis)
	return 0
}

func jevtProgress(ts time.Time, evt ggja.Obj) jePost {
	stateLock.Lock()
	defer stateLock.Unlock()
	theCmdr.Ranks.Imps.Progress = evt.Int("Empire", theCmdr.Ranks.Imps.Progress)
	theCmdr.Ranks.Feds.Progress = evt.Int("Federation", theCmdr.Ranks.Feds.Progress)
	theCmdr.Ranks.Combat.Progress = evt.Int("Combat", theCmdr.Ranks.Combat.Progress)
	theCmdr.Ranks.Trade.Progress = evt.Int("Trade", theCmdr.Ranks.Trade.Progress)
	theCmdr.Ranks.Explore.Progress = evt.Int("Explore", theCmdr.Ranks.Explore.Progress)
	theCmdr.Ranks.CQC.Progress = evt.Int("CQC", theCmdr.Ranks.CQC.Progress)
	return 0
}

func jevtLocation(ts time.Time, evt ggja.Obj) jePost {
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
	eddnSendJournal(theEddn, ts, evt, ssys)
	return jePostHdr
}

func jevtFsdJump(ts time.Time, evt ggja.Obj) jePost {
	stateLock.Lock()
	defer stateLock.Unlock()
	sysNm := evt.MStr("StarSystem")
	coos := evt.MArr("StarPos")
	ssys := sysByName(sysNm, coos)
	theCmdr.Loc.SysId = ssys.Id
	theCmdr.Loc.LocId = 0
	theCmdr.Loc.Docked = false
	theCmdr.Jumps.Add(ssys.Id, ts)
	eddnSendJournal(theEddn, ts, evt, ssys)
	return jePostHdr
}

func jevtScan(ts time.Time, evt ggja.Obj) jePost {
	const jeScanAction = 0
	// TODO update cmdr location
	if theCmdr.Loc.SysId > 0 {
		ssys, err := theGalaxy.GetSystem(theCmdr.Loc.SysId)
		if err != nil {
			panic(err)
		}
		if ssys == nil {
			return jeScanAction
		}
		eddnSendJournal(theEddn, ts, evt, ssys)
	}
	return jeScanAction
}

func jevtShipyardBuy(ts time.Time, evt ggja.Obj) jePost {
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
	nameMaps.ShipType.PickL10n(evt, "ShipType", "ShipType_Localised")
	return jePostHdr
}

func jevtShipyardNew(ts time.Time, evt ggja.Obj) jePost {
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
	nameMaps.ShipType.PickL10n(evt, "ShipType", "ShipType_Localised")
	return jePostHdr
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

func jevtShipyardSell(ts time.Time, evt ggja.Obj) jePost {
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
	nameMaps.ShipType.PickL10n(evt, "ShipType", "ShipType_Localised")
	return jePostHdr
}

func jevtShipyardSwap(ts time.Time, evt ggja.Obj) jePost {
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
	nameMaps.ShipType.PickL10n(evt, "ShipType", "ShipType_Localised")
	return jePostHdr
}

func jevtShipTargeted(ts time.Time, evt ggja.Obj) jePost {
	nameMaps.ShipType.PickL10n(evt, "Ship", "Ship_Localised")
	return 0
}

func jevtShipyardTransfer(ts time.Time, evt ggja.Obj) jePost {
	sid := evt.MInt("ShipID")
	stateLock.Lock()
	defer stateLock.Unlock()
	if theCmdr.InShip == sid {
		log.Logf(l.Lwarn, "commander in transferred ship %d", sid)
		theCmdr.InShip = cmdr.NoShip
	}
	ship := theCmdr.MustHaveShip(sid, evt.MStr("ShipType"))
	ship.BerthLoc = theCmdr.Loc.LocId
	return 0
}

func jevtShutdown(ts time.Time, evt ggja.Obj) jePost {
	stateLock.RLock()
	defer stateLock.RUnlock()
	switchToCommander("")
	return jePostReload
}

func jevtStartJump(ts time.Time, evt ggja.Obj) jePost {
	jty := evt.MStr("JumpType")
	if jty == "Supercruise" {
		stateLock.Lock()
		defer stateLock.Unlock()
		theCmdr.Loc.LocId = 0
		theCmdr.Loc.Docked = false
	}
	return jePostHdr
}

func jevtSupercruiseExit(ts time.Time, evt ggja.Obj) jePost {
	// TODO NYI
	return 0
}
