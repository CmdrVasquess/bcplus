package core

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/CmdrVasquess/BCplus/webui"

	"git.fractalqb.de/fractalqb/ggja"
	log "git.fractalqb.de/fractalqb/qbsllm"
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
	lgr.Debuga("spool journal events after `time`", startAfter)
	rddir, err := os.Open(jdir)
	if err != nil {
		lgr.Errora("fail to scan journal-dir: `err`", err)
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
	lgr.Infoa("reading missed events from `file`", jfnm)
	jf, err := os.Open(jfnm)
	if err != nil {
		lgr.Errora("cannot open journal: `err`", err)
		return
	}
	defer jf.Close()
	scn := bufio.NewScanner(jf)
	for scn.Scan() {
		journalEvent(scn.Bytes())
	}
}

type jeActn struct {
	hdlr    func(ts time.Time, evt ggja.Obj) webui.UIUpdate
	backlog bool
}

var jEvtMatCat = map[string]cmdr.MatKind{
	"Raw":          cmdr.MatRaw,
	"Manufactured": cmdr.MatMan,
	"Encoded":      cmdr.MatEnc,
}

var jEventHdl = map[string]jeActn{
	"Commander":         jeActn{jevtCommander, false},
	"DiscoveryScan":     jeActn{jevtDiscoveryScan, true},
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

func journalEvent(jLine []byte) (wuiupd webui.UIUpdate) {
	eJson := make(ggja.GenObj)
	err := json.Unmarshal(jLine, &eJson)
	if err != nil {
		lgr.Errora("cannot parse journal event: `error`@`line`", err, string(jLine))
		return 0
	}
	defer func() {
		if p := recover(); p != nil {
			lgr.Errora("recover journal `panic`", p)
		}
	}()
	evt := ggja.Obj{Bare: eJson}
	ets := evt.MTime("timestamp")
	enm := evt.Str("event", "")
	if len(enm) == 0 {
		lgr.Errora("no event name in journal `event`", string(jLine))
		return 0
	}
	if ets.Before(bcpState.LastEDEvent) {
		lgr.Tracea("ignore historic event `name` `at` <= `start`",
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
		return 0
	}
	if theCmdr != nil {
		jEventMacro(enm, theCmdr.JStatFlags)
	}
	acnt, ok := jEventHdl[enm]
	if ok {
		if theCmdr == nil && acnt.backlog {
			lgr.Debuga("putting `event` to backlog", enm)
			jevtBacklog = append(jevtBacklog, eJson)
		} else {
			if jevtSpooling {
				lgr.Debuga("spooling to `event` handler", enm)
			} else {
				lgr.Debuga("dispatch to `event` handler", enm)
			}
			if len(jevtBacklog) > 0 {
				var nbl []ggja.GenObj
				for _, jEvt := range jevtBacklog {
					evt := ggja.Obj{Bare: jEvt}
					ets := evt.MTime("timestamp")
					enm := evt.MStr("event")
					blActn, _ := jEventHdl[enm]
					lgr.Debuga("dispatch from backlog to `event` handler", enm)
					wuiupd |= blActn.hdlr(ets, ggja.Obj{Bare: jEvt})
				}
				jevtBacklog = nbl
			}
			wuiupd |= acnt.hdlr(ets, evt)
			bcpState.LastEDEvent = ets
		}
	} else {
		lgr.Debuga("no handler for `event`", enm)
	}
	return wuiupd
}

func sysByName(name string, coos *ggja.Arr) (res *galaxy.System) {
	var err error
	if coos == nil {
		res, err = theGalaxy.MustSystem(name)
	} else {
		res, err = theGalaxy.MustSystemCoos(name,
			coos.MF64(0), coos.MF64(1), coos.MF64(2))
	}
	if err != nil {
		lgr.Panica("`err`", err)
	}
	return res
}

func partFromSys(sys *galaxy.System, typ galaxy.PartType, partName string) *galaxy.SysPart {
	res, err := sys.MustPart(typ, partName)
	if err != nil {
		lgr.Panica("`err`", err)
	}
	return res
}

func jevtFileheader(ts time.Time, evt ggja.Obj) webui.UIUpdate {
	stateLock.Lock()
	defer stateLock.Unlock()
	lang := evt.MStr("language")
	locale, dom := nameMaps.Lang.Map(nameMaps.LangEd, lang, nameMaps.LangLoc)
	nameMaps.Save()
	if dom >= 0 {
		nameMaps.Load(resDir, FlagDDir, locale)
	} else {
		nameMaps.Load(resDir, FlagDDir, common.DefaultLang)
	}
	switchToCommander("")
	return webui.UIReload
}

func jevtCommander(ts time.Time, evt ggja.Obj) webui.UIUpdate {
	name := evt.MStr("Name")
	stateLock.Lock()
	defer stateLock.Unlock()
	switchToCommander(name)
	return webui.UIReload
}

func jevtDiscoveryScan(ts time.Time, evt ggja.Obj) webui.UIUpdate {
	stateLock.Lock()
	defer stateLock.Unlock()
	ssys, err := theGalaxy.GetSystem(theCmdr.Loc.SysId)
	if err != nil {
		panic(err)
	}
	scnBno := evt.MInt("Bodies")
	if ssys.BodyNo != scnBno {
		ssys.BodyNo = scnBno
		_, err := theGalaxy.PutSystem(ssys)
		if err != nil {
			panic(err)
		}
	}
	return 0
}

func jevtDocked(ts time.Time, evt ggja.Obj) webui.UIUpdate {
	stateLock.Lock()
	defer stateLock.Unlock()
	ssys := sysByName(evt.MStr("StarSystem"), nil)
	port, err := ssys.MustPart(galaxy.Port, evt.MStr("StationName"))
	theCmdr.Loc.SysId = ssys.Id
	theCmdr.Loc.LocId = port.Id
	theCmdr.Loc.Docked = true
	if err != nil {
		lgr.Panica("`err`", err)
	}
	eddnSendJournal(theEddn, ts, evt, ssys)
	return webui.UIHdr
}

func jevtUndocked(ts time.Time, evt ggja.Obj) webui.UIUpdate {
	stateLock.Lock()
	defer stateLock.Unlock()
	// TODO compare station/port; problem unreliable system id
	theCmdr.Loc.Docked = false
	return 0
}

func jevtLoadGame(ts time.Time, evt ggja.Obj) webui.UIUpdate {
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
	return webui.UIReload
}

func jevtModSzNCls(item string) (name string, size, class int) {
	var err error
	szIdx := strings.LastIndex(item, "_Size")
	if szIdx < 0 {
		return item, 0, 0
	}
	clsIdx := strings.LastIndex(item, "_Class")
	if clsIdx < 0 {
		return item, 0, 0
	}
	size, err = strconv.Atoi(item[szIdx+5 : clsIdx])
	if err != nil {
		panic(err)
	}
	class, err = strconv.Atoi(item[clsIdx+6:])
	if err != nil {
		panic(err)
	}
	name = item[:szIdx]
	return name, size, class
}

func jevtSlot2Mod(sm ggja.Obj) (kind cmdr.SlotKind, mdl *cmdr.Module) {
	item := sm.MStr("Item")
	switch slot := sm.MStr("Slot"); slot {
	case "Armour":
		c, _ := strconv.Atoi(item[len(item)-1:])
		name := item[:len(item)-7]
		if sep := strings.IndexRune(name, '_'); sep > 0 {
			name = name[sep+1:]
		}
		mdl = &cmdr.Module{
			Kind:  cmdr.CoreModule,
			Slot:  0,
			Name:  name,
			Class: c,
		}
	case "PowerPlant":
		n, s, c := jevtModSzNCls(item)
		mdl = &cmdr.Module{
			Kind:  cmdr.CoreModule,
			Slot:  1,
			Name:  n[4:],
			Size:  s,
			Class: c,
		}
	case "MainEngines":
		n, s, c := jevtModSzNCls(item)
		mdl = &cmdr.Module{
			Kind:  cmdr.CoreModule,
			Slot:  2,
			Name:  n[4:],
			Size:  s,
			Class: c,
		}
	case "FrameShiftDrive":
		n, s, c := jevtModSzNCls(item)
		mdl = &cmdr.Module{
			Kind:  cmdr.CoreModule,
			Slot:  3,
			Name:  n[4:],
			Size:  s,
			Class: c,
		}
	case "LifeSupport":
		n, s, c := jevtModSzNCls(item)
		mdl = &cmdr.Module{
			Kind:  cmdr.CoreModule,
			Slot:  4,
			Name:  n[4:],
			Size:  s,
			Class: c,
		}
	case "PowerDistributor":
		n, s, c := jevtModSzNCls(item)
		mdl = &cmdr.Module{
			Kind:  cmdr.CoreModule,
			Slot:  5,
			Name:  n[4:],
			Size:  s,
			Class: c,
		}
	case "Radar":
		n, s, c := jevtModSzNCls(item)
		mdl = &cmdr.Module{
			Kind:  cmdr.CoreModule,
			Slot:  6,
			Name:  n[4:],
			Size:  s,
			Class: c,
		}
	case "FuelTank":
		n, s, c := jevtModSzNCls(item)
		mdl = &cmdr.Module{
			Kind:  cmdr.CoreModule,
			Slot:  7,
			Name:  n[4:],
			Size:  s,
			Class: c,
		}
	case "PlanetaryApproachSuite":
		return cmdr.OptModule, nil
	case "WeaponColour":
		return cmdr.OptModule, nil
	case "EngineColour":
		return cmdr.OptModule, nil
	case "VesselVoice":
		return cmdr.OptModule, nil
	case "ShipCockpit":
		return cmdr.CoreModule, nil
	case "CargoHatch":
		return cmdr.CoreModule, nil
	default:
		if strings.HasPrefix(slot, "Slot") {
			slotNo, _ := strconv.Atoi(slot[4:6])
			n, s, c := jevtModSzNCls(item)
			mdl = &cmdr.Module{
				Kind:  cmdr.OptModule,
				Slot:  slotNo,
				Name:  n[4:],
				Size:  s,
				Class: c,
			}
		} else if pos := strings.LastIndex(slot, "Hardpoint"); pos >= 0 {
			slotNo, _ := strconv.Atoi(slot[pos+9:])
			mdl = &cmdr.Module{
				Kind: cmdr.Hardpoint,
				Slot: slotNo,
				Name: item,
			}
		} else if strings.HasPrefix(slot, "Decal") {
			return cmdr.OptModule, nil
		} else if strings.HasPrefix(slot, "ShipName") {
			return cmdr.OptModule, nil
		} else if strings.HasPrefix(slot, "Bobble") {
			return cmdr.OptModule, nil
		} else if strings.HasPrefix(slot, "PaintJob") {
			return cmdr.OptModule, nil
		} else {
			lgr.Warna("unknown module `slot`", slot)
			return cmdr.OptModule, nil
		}
	}
	return mdl.Kind, mdl
}

func jevtLoadout(ts time.Time, evt ggja.Obj) webui.UIUpdate {
	stateLock.Lock()
	defer stateLock.Unlock()
	if theCmdr.JStatFlags&statInMainShip != statInMainShip {
		lgr.Debug(log.Str("ignore loadout for non-mainship"))
		return 0
	}
	ship := theCmdr.MustHaveShip(evt.MInt("ShipID"), evt.MStr("Ship"))
	theCmdr.InShip = ship.Id
	ship.BerthLoc = 0
	ship.Ident = evt.Str("ShipIdent", ship.Ident)
	ship.Name = evt.Str("ShipName", ship.Name)
	ship.Health = evt.F64("Health", ship.Health)
	ship.Rebuy = evt.Int("Rebuy", ship.Rebuy)
	ship.HullValue = evt.Int("HullValue", ship.HullValue)
	ship.ModuleValue = evt.Int("ModulesValue", ship.ModuleValue)
	if modls := evt.Arr("Modules"); modls != nil {
		ship.Hardpoints = nil
		ship.Utilities = nil
		ship.CoreModules = nil
		ship.OptModules = nil
		for _, jm := range modls.Bare {
			s := ggja.Obj{Bare: jm.(ggja.GenObj), OnError: evt.OnError}
			kind, mod := jevtSlot2Mod(s)
			if mod == nil {
				continue
			}
			switch kind {
			case cmdr.Hardpoint:
				ship.Hardpoints = append(ship.Hardpoints, mod)
			case cmdr.Utility:
				ship.Utilities = append(ship.Utilities, mod)
			case cmdr.CoreModule:
				ship.CoreModules = append(ship.CoreModules, mod)
			case cmdr.OptModule:
				ship.OptModules = append(ship.OptModules, mod)
			}
		}
		sort.Slice(ship.Hardpoints, func(i, j int) bool {
			return ship.Hardpoints[i].Slot < ship.Hardpoints[j].Slot
		})
		sort.Slice(ship.Utilities, func(i, j int) bool {
			return ship.Utilities[i].Slot < ship.Utilities[j].Slot
		})
		sort.Slice(ship.CoreModules, func(i, j int) bool {
			return ship.CoreModules[i].Slot < ship.CoreModules[j].Slot
		})
		sort.Slice(ship.OptModules, func(i, j int) bool {
			return ship.OptModules[i].Slot < ship.OptModules[j].Slot
		})
	}
	return webui.UIHdr
}

func akkuMats(cat string, mat string, d int) {
	catId, ok := jEvtMatCat[cat]
	if !ok {
		lgr.Warna("unknown material `category`", cat)
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

func jevtMaterialCollected(ts time.Time, evt ggja.Obj) webui.UIUpdate {
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

func jevtMaterials(ts time.Time, evt ggja.Obj) webui.UIUpdate {
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

func jevtMissionAbandoned(ts time.Time, evt ggja.Obj) webui.UIUpdate {
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
	theCmdr.MissPath = nil
	return webui.UIMissions
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

func jevtMissionAccepted(ts time.Time, evt ggja.Obj) webui.UIUpdate {
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
	if dst := evt.Str("DestinationSystem", ""); dst != "" {
		dests := splitMultiDestination(dst)
		var (
			sys *galaxy.System
			err error
		)
		for _, d := range dests {
			sys, err = theGalaxy.MustSystem(d)
			if err != nil {
				lgr.Panica("`err`", err)
			}
			mssn.Dests = append(mssn.Dests, sys.Id)
			if !galaxy.V3dValid(sys.Coos) {
				lgr.Warna("system `id` `name` without coos", sys.Id, sys.Name)
			}
		}
	}
	theCmdr.MissPath = nil
	return webui.UIMissions
}

func jevtMissionCompleted(ts time.Time, evt ggja.Obj) webui.UIUpdate {
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
				lgr.Warna("illegal mission reputation trend: `rep`", rep)
			}
		}
	}
	theCmdr.MissPath = nil
	return webui.UIMissions
}

func jevtMissions(ts time.Time, evt ggja.Obj) webui.UIUpdate {
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
	theCmdr.MissPath = nil
	return 0
}

func jevtRank(ts time.Time, evt ggja.Obj) webui.UIUpdate {
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

func jevtReputation(ts time.Time, evt ggja.Obj) webui.UIUpdate {
	stateLock.Lock()
	defer stateLock.Unlock()
	theCmdr.Rep.Imps = evt.F32("Empire", theCmdr.Rep.Imps)
	theCmdr.Rep.Feds = evt.F32("Federation", theCmdr.Rep.Feds)
	theCmdr.Rep.Allis = evt.F32("Alliance", theCmdr.Rep.Allis)
	return 0
}

func jevtProgress(ts time.Time, evt ggja.Obj) webui.UIUpdate {
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

func jevtLocation(ts time.Time, evt ggja.Obj) webui.UIUpdate {
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
				lgr.Panica("`err`", err)
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
	return webui.UIHdr
}

func jevtFsdJump(ts time.Time, evt ggja.Obj) webui.UIUpdate {
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
	return webui.UIHdr
}

func jevtBodyType(evt ggja.Obj) galaxy.PartType {
	if tmp := evt.Str("StarType", ""); tmp != "" {
		return galaxy.Star
	}
	if tmp := evt.MStr("BodyName"); strings.Contains(tmp, "Belt") {
		return galaxy.Belt
	}
	return galaxy.Planet
}

func jevtScan(ts time.Time, evt ggja.Obj) webui.UIUpdate {
	const jeScanAction = 0
	if theCmdr.Loc.SysId <= 0 {
		return 0
	}
	ssys, err := theGalaxy.GetSystem(theCmdr.Loc.SysId)
	if err != nil {
		panic(err)
	}
	if ssys == nil {
		return jeScanAction
	}
	eddnSendJournal(theEddn, ts, evt, ssys)
	// currently gnore "ScanType"
	locNm := galaxy.LocalName(ssys.Name, evt.MStr("BodyName"))
	part, err := ssys.MustPart(jevtBodyType(evt), locNm)
	if err != nil {
		panic(err)
	}
	part.FromCenter = evt.F32("DistanceFromArrivalLS", -1.0)
	part, err = theGalaxy.PutSysPart(part)
	if err != nil {
		panic(err)
	}
	if mats := evt.Arr("Materials"); mats != nil {
		// TODO
	}
	return jeScanAction
}

func jevtShipyardBuy(ts time.Time, evt ggja.Obj) webui.UIUpdate {
	storeId := evt.MInt("StoreShipID")
	if storeId != theCmdr.InShip {
		lgr.Warna("inconsistent ids for current ship `bc+` / `event`",
			theCmdr.InShip, storeId)
	}
	stateLock.Lock()
	defer stateLock.Unlock()
	ship := theCmdr.MustHaveShip(storeId, evt.MStr("ShipType"))
	ship.BerthLoc = theCmdr.Loc.LocId
	theCmdr.InShip = cmdr.NoShip
	nameMaps.ShipType.PickL10n(evt, "ShipType", "ShipType_Localised")
	return webui.UIHdr | webui.UIShips
}

func jevtShipyardNew(ts time.Time, evt ggja.Obj) webui.UIUpdate {
	ship := &cmdr.Ship{
		Id:     evt.MInt("NewShipID"),
		Type:   evt.MStr("ShipType"),
		Bought: ts,
	}
	stateLock.Lock()
	defer stateLock.Unlock()
	if theCmdr.InShip > cmdr.NoShip {
		lgr.Warna("commander in `ship` when new ship arrives", theCmdr.InShip)
		if s2s, ok := theCmdr.Ships[theCmdr.InShip]; ok {
			s2s.BerthLoc = theCmdr.Loc.LocId
		}
	}
	theCmdr.Ships[ship.Id] = ship
	theCmdr.InShip = ship.Id
	nameMaps.ShipType.PickL10n(evt, "ShipType", "ShipType_Localised")
	return webui.UIHdr | webui.UIShips
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
		lgr.Warna("cannot save sold ship `id` to `file`", ship.Id, fnm)
		return
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	err = enc.Encode(ship)
}

func jevtShipyardSell(ts time.Time, evt ggja.Obj) webui.UIUpdate {
	sid := evt.MInt("SellShipID")
	stateLock.Lock()
	defer stateLock.Unlock()
	if theCmdr.InShip == sid {
		lgr.Warna("commander in sold ship `id`", sid)
		theCmdr.InShip = cmdr.NoShip
	}
	if ship, ok := theCmdr.Ships[sid]; ok && ship != nil {
		go saveSoldShip(theCmdr.Name, ship)
	}
	delete(theCmdr.Ships, sid)
	nameMaps.ShipType.PickL10n(evt, "ShipType", "ShipType_Localised")
	return webui.UIHdr | webui.UIShips
}

func jevtShipyardSwap(ts time.Time, evt ggja.Obj) webui.UIUpdate {
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
	return webui.UIHdr | webui.UIShips
}

func jevtShipTargeted(ts time.Time, evt ggja.Obj) webui.UIUpdate {
	nameMaps.ShipType.PickL10n(evt, "Ship", "Ship_Localised")
	return 0
}

func jevtShipyardTransfer(ts time.Time, evt ggja.Obj) webui.UIUpdate {
	sid := evt.MInt("ShipID")
	stateLock.Lock()
	defer stateLock.Unlock()
	if theCmdr.InShip == sid {
		lgr.Warna("commander in transferred ship `id`", sid)
		theCmdr.InShip = cmdr.NoShip
	}
	ship := theCmdr.MustHaveShip(sid, evt.MStr("ShipType"))
	ship.BerthLoc = theCmdr.Loc.LocId
	return 0
}

func jevtShutdown(ts time.Time, evt ggja.Obj) webui.UIUpdate {
	stateLock.RLock()
	defer stateLock.RUnlock()
	switchToCommander("")
	return webui.UIReload
}

func jevtStartJump(ts time.Time, evt ggja.Obj) webui.UIUpdate {
	jty := evt.MStr("JumpType")
	if jty == "Supercruise" {
		stateLock.Lock()
		defer stateLock.Unlock()
		theCmdr.Loc.LocId = 0
		theCmdr.Loc.Docked = false
	}
	return webui.UIHdr
}

func jevtSupercruiseExit(ts time.Time, evt ggja.Obj) webui.UIUpdate {
	// TODO NYI
	return 0
}
