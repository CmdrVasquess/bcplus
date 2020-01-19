package app

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"git.fractalqb.de/fractalqb/ggja"
	"git.fractalqb.de/fractalqb/namemap"
	"git.fractalqb.de/fractalqb/qbsllm"
	"github.com/CmdrVasquess/bcplus/internal/ship"
	"github.com/CmdrVasquess/watched"
)

var (
	jelog    = qbsllm.New(qbsllm.Lnormal, "e-journal", nil, nil)
	jelogCfg = qbsllm.Config(jelog)
)

type journalHandler func(t time.Time, evt ggja.Obj) Change

var jHdlrs = map[string]journalHandler{
	"ApproachBody":        jeApproachBody,
	"ApproachSettlement":  jeApproachSettlement,
	"Commander":           jeCommander,
	"EndCrewSession":      jeEndCrewSession,
	"FSDJump":             jeFsdJump,
	"FSSAllBodiesFound":   jeFSSAllBodiesFound,
	"FSSDiscoveryScan":    jeFSSDiscoveryScan,
	"FSSSignalDiscovered": jeFSSSignalDiscovered,
	"JoinACrew":           jeJoinACrew,
	"LeaveBody":           jeLeaveBody,
	"LoadGHame":           jeLoadGame,
	"Loadout":             jeLoadout,
	"Location":            jeLocation,
	"Materials":           jeMaterials,
	"NewCommander":        jeCommander,
	"ReceiveText":         jeReceiveText,
	"SAASignalsFound":     jeSAASignalsFound,
	"Scan":                jeScan,
	"Screenshot":          jeScreenshot,
	"Shutdown":            jsShutdown,
	"StartJump":           jeStartJump,
	"SupercruiseExit":     jeSupercruiseExit,
	"Touchdown":           jeTouchdown,
	/*
		Docked/Undocked
		Liftoff/Touchdown
		SupercruiseEntry/Exit
		ShipyardSwap …?
		VehicleSwitch
	*/
}

func jEvtType(str string) (timestamp time.Time, event string) {
	var err error
	idx := strings.Index(str, `"timestamp":`)
	if idx < 0 {
		panic("no timestamp in event")
	}
	val := str[idx+13 : idx+33]
	timestamp, err = time.Parse(time.RFC3339, val)
	if err != nil {
		panic(err)
	}
	str = str[idx+35:]
	idx = strings.Index(str, `"event":`)
	if idx < 0 {
		panic("no event type in event")
	}
	str = str[idx+9:]
	idx = strings.IndexByte(str, '"')
	if idx < 0 {
		panic("cannot find end of event type")
	}
	return timestamp, str[:idx]
}

func journalEvent(str string, useTs bool) (chg Change, uts bool) {
	defer recoverEvent("journal", str)
	jEvtType(str)
	ts, etype := jEvtType(str)
	switch {
	case jelog.Logs(qbsllm.Ltrace):
		jelog.Traces(str)
	case jelog.Logs(qbsllm.Ldebug):
		e := str
		if len(e) > 80 {
			e = e[:77] + "[…]"
		}
		jelog.Debuga("`event` `at` `is`", etype, ts, e)
	}
	// Cope with non-monotonic timestamps in logs after summer->winter time
	// switch in 2019:
	switch etype {
	case "Fileheader":
		useTs = false
		jelog.Tracea("`event` switches timestamp handling off")
	case "Commander":
		useTs = true
		jelog.Tracea("`event` switches timestamp handling on")
	}
	if useTs {
		if ts.Before(App.LastEvent) {
			// TODO At game start events tend to have same timestamp (bursts)
			log.Tracea("drop `old` `journal event` happened `before`",
				ts,
				etype,
				App.LastEvent)
			return 0, useTs
		} else {
			App.LastEvent = ts
		}
	}
	hdlr := jHdlrs[etype]
	if hdlr == nil {
		log.Tracea("no handler for `journal event` `at`", etype, ts)
	} else {
		log.Debuga("handle journal `event`", etype)
		dec := json.NewDecoder(strings.NewReader(str))
		je := make(ggja.GenObj)
		err := dec.Decode(&je)
		if err != nil {
			panic(err)
		}
		evt := ggja.Obj{Bare: je}
		chg = hdlr(ts, evt)
	}
	return chg, useTs
}

func jeCommander(t time.Time, evt ggja.Obj) Change {
	fid := evt.MStr("FID")
	nm := evt.MStr("Name")
	writeState(noErr(func() {
		cmdr.switchTo(fid, nm)
	}))
	return ChgCmdr
}

func jeReceiveText(t time.Time, evt ggja.Obj) Change {
	if toSpeak != nil {
		msg := evt.Str("Message_Localised", "")
		if msg == "" {
			msg = evt.MStr("Message")
		}
		from := evt.Str("From_Localised", "")
		if from == "" {
			from = evt.MStr("From")
		}
		msg = "From " + from + ": " + msg
		chn := evt.MStr("Channel")
		if chn == "squadron" {
			dispatchVoice(chn, 1, msg)
		} else {
			dispatchVoice(chn, 0, msg)
		}
	}
	return 0
}

func jeSAASignalsFound(t time.Time, evt ggja.Obj) Change {
	if toSpeak == nil {
		return 0
	}
	sigs := evt.Arr("Signals")
	if sigs.Len() > 0 {
		var sb strings.Builder
		fmt.Fprintf(&sb, "Found")
		for i, rsig := range sigs.Bare {
			sig := ggja.Obj{Bare: rsig.(ggja.GenObj), OnError: evt.OnError}
			if i > 0 {
				sb.WriteString(" and")
			}
			fmt.Fprintf(&sb, " %d %s signals",
				sig.MInt("Count"),
				sig.MStr("Type_Localised"))
		}
		fmt.Fprintf(&sb, " on %s.", evt.MStr("BodyName"))
		dispatchVoice(ChanJEvt, 0, sb.String())
	}
	return 0
}

func jeScan(t time.Time, evt ggja.Obj) Change {
	sysnm := evt.MStr("StarSystem")
	bdynm := evt.MStr("BodyName")
	if bdynm != sysnm && strings.HasPrefix(bdynm, sysnm) {
		bdynm = strings.TrimSpace(bdynm[len(sysnm):])
	}
	b := &InSysBody{
		Id:        evt.MInt("BodyID"),
		Name:      bdynm,
		Dist:      evt.MF32("DistanceFromArrivalLS"),
		R:         evt.F32("Radius", 0),
		Grav:      evt.F32("SurfaceGravity", 0),
		Temp:      evt.F32("SurfaceTemperature", 0),
		Volcano:   evt.Str("Volcanism", ""),
		Land:      evt.Bool("Landable", false),
		TidalLock: evt.Bool("TidalLock", false),
		Disco:     evt.MBool("WasDiscovered"),
		Mapd:      evt.MBool("WasMapped"),
	}
	writeState(noErr(func() {
		inSysInfo.Bodies = append(inSysInfo.Bodies, b)
	}))
	if toSpeak != nil && (evt.MStr("ScanType") != "AutoScan" || evt.MInt("BodyID") == 0) {
		var msg string
		if b.Disco {
			if b.Mapd {
				msg = fmt.Sprintf("%s: mapped.", bdynm)
			} else {
				msg = fmt.Sprintf("%s: not mapped!", bdynm)
			}
		} else {
			msg = fmt.Sprintf("%s: not discovered!", bdynm)
		}
		dispatchVoice(ChanJEvt, 0, msg)
	}
	return WuiUpInSys
}

func jsShutdown(t time.Time, evt ggja.Obj) Change {
	writeState(noErr(func() {
		cmdr.switchTo("", "")
		App.save()
	}))
	return ChgCmdr
}

func jeTouchdown(t time.Time, evt ggja.Obj) Change {
	var chg Change
	writeState(noErr(func() {
		chg |= cmdr.Loc.SetMode(Parked)
		if nd := evt.Str("NearestDestination_Localised", ""); nd == "" {
			chg |= cmdr.Loc.SetRef(Planet)
		} else {
			chg |= cmdr.Loc.SetRef(RefUndef)
			chg |= cmdr.Loc.SetRefNm(nd)
		}
	}))
	return chg
}

func jeParseMats(mats map[string]MatState, jeMats *ggja.Arr, nms *namemap.NameMap) {
	ed2int := nms.From("ed", false).To(true)
	for _, jem := range jeMats.Bare {
		jeMat := ggja.Obj{Bare: jem.(ggja.GenObj), OnError: jeMats.OnError}
		edNm := jeMat.MStr("Name")
		intNm, toDom := ed2int.Map(edNm)
		if toDom < 0 {
			log.Warna("cannot map `ed-material` to internal key", edNm)
			continue
		}
		count := jeMat.MInt("Count")
		mat := cmdr.Mats[intNm]
		mat.Have = count
		cmdr.Mats[intNm] = mat
	}
}

func jeMaterials(t time.Time, evt ggja.Obj) Change {
	if !cmdr.isVoid() {
		mats := make(map[string]MatState)
		jeParseMats(mats, evt.MArr("Raw"), nmRawMat)
		jeParseMats(mats, evt.MArr("Manufactured"), nmManMat)
		jeParseMats(mats, evt.MArr("Encoded"), nmEncMat)
		cmdr.Mats = mats
	}
	matfnm := filepath.Join(cmdrDir(cmdr.Fid), "mats.json")
	wr, err := os.Create(matfnm)
	if err != nil {
		log.Panice(err)
	}
	defer wr.Close()
	enc := json.NewEncoder(wr)
	enc.SetIndent("", "  ")
	err = enc.Encode(evt.Bare)
	if err != nil {
		log.Errore(err)
	}
	return 0
}

func jeLoadout(t time.Time, evt ggja.Obj) (chg Change) {
	shipId := evt.MInt("ShipID")
	shp := ship.TheShips.Load(evt.MInt("ShipID"), evt.MStr("Ship"))
	if shp.Type.ShipType.Refine(evt) {
		err := ship.TheTypes.Save(shp.Type.ShipType)
		log.Errore(err)
	}
	shp.Update(evt)
	writeState(noErr(func() {
		if cmdr.isVoid() {
			log.Errors("loadout event without commander")
			return
		}
		cmdr.Ship.Ship = shp
		chg = ChgShip
	}))
	shipsDir := filepath.Join(cmdrDir(cmdr.Fid), "ships")
	if _, err := os.Stat(shipsDir); os.IsNotExist(err) {
		err = os.MkdirAll(shipsDir, 0777)
		if err != nil {
			log.Panice(err)
		}
	}
	shpFNm := filepath.Join(shipsDir, fmt.Sprintf("ship-%d.json", shipId))
	wr, err := os.Create(shpFNm)
	if err != nil {
		log.Errore(err)
	}
	defer wr.Close()
	enc := json.NewEncoder(wr)
	enc.SetIndent("", "  ")
	err = enc.Encode(evt.Bare)
	if err != nil {
		log.Errore(err)
	}
	return chg
}

func jeLocation(t time.Time, evt ggja.Obj) Change {
	var chg Change
	writeState(noErr(func() {
		inSysInfo.reset()
		chg |= cmdr.Loc.SetSys(
			evt.MUint64("SystemAddress"),
			evt.MStr("StarSystem"),
		)
		if evt.MBool("Docked") {
			chg |= cmdr.Loc.SetMode(Parked)
		} else {
			chg |= cmdr.Loc.SetMode(Move)
		}
		if ref := evt.Str("StationName", ""); len(ref) > 0 {
			chg |= cmdr.Loc.SetRefNm(ref)
			rty := evt.MStr("StationType")
			if ty := jeStnTypeMap[rty]; ty == RefUndef {
				log.Warna("unknown `station type` in location event", rty)
				chg |= cmdr.Loc.SetRef(RefUndef)
			} else {
				chg |= cmdr.Loc.SetRef(ty)
			}
		} else if ref := evt.Str("Body", ""); len(ref) > 0 {
			chg |= cmdr.Loc.SetRefNm(ref)
			rty := evt.MStr("BodyType")
			if ty := jeBodyTypeMap[rty]; ty == RefUndef {
				log.Warna("unknown `body type` in location event", t)
				chg |= cmdr.Loc.SetRef(RefUndef)
			} else {
				chg |= cmdr.Loc.SetRef(ty)
			}
		}
	}))
	return chg
}

func jeApproachBody(t time.Time, evt ggja.Obj) Change {
	var chg Change
	writeState(noErr(func() {
		chg |= cmdr.Loc.SetRef(RefUndef)
		chg |= cmdr.Loc.SetRefNm(evt.MStr("Body"))
	}))
	return chg
}

func jeFsdJump(t time.Time, evt ggja.Obj) Change {
	var chg Change
	writeState(noErr(func() {
		chg |= cmdr.Loc.SetSys(
			evt.MUint64("SystemAddress"),
			evt.MStr("StarSystem"),
		)
		chg |= cmdr.Loc.SetRef(jeBodyTypeMap[evt.MStr("BodyType")])
		chg |= cmdr.Loc.SetRefNm(evt.MStr("Body"))
		chg |= cmdr.Loc.SetMode(Cruise)
		inSysInfo.reset()
	}))
	return chg | WuiUpInSys
}

func jeFSSAllBodiesFound(t time.Time, evt ggja.Obj) Change {
	inSysInfo.BodyNum = evt.MInt("Count")
	return WuiUpInSys
}

func jeFSSDiscoveryScan(t time.Time, evt ggja.Obj) Change {
	inSysInfo.BodyNum = evt.MInt("BodyCount")
	inSysInfo.MiscNum = evt.MInt("NonBodyCount")
	if toSpeak != nil {
		var msg string
		if inSysInfo.BodyNum > 0 {
			if inSysInfo.MiscNum > 0 {
				msg = fmt.Sprintf(
					"Discovered %d bodies and %d signals",
					inSysInfo.BodyNum,
					inSysInfo.MiscNum,
				)
			} else {
				msg = fmt.Sprintf("Discovered %d bodies", inSysInfo.BodyNum)
			}
		} else if inSysInfo.MiscNum > 0 {
			msg = fmt.Sprintf("Discovered %d signals", inSysInfo.MiscNum)
		}
		dispatchVoice(ChanJEvt, 0, msg)
	}
	return WuiUpInSys
}

func jeFSSSignalDiscovered(t time.Time, evt ggja.Obj) Change {
	nm := evt.Str("SignalName_Localised", "")
	writeState(noErr(func() {
		if nm == "" {
			inSysInfo.addSignal(evt.MStr("SignalName"))
		} else {
			inSysInfo.addSignal(nm)
		}
	}))
	if strings.Index(strings.ToLower(nm), "notable") >= 0 {
		dispatchVoice(ChanJEvt, 1, nm)
	}
	return WuiUpInSys
}

func jeStartJump(t time.Time, evt ggja.Obj) Change {
	var chg Change
	switch evt.MStr("JumpType") {
	case "Hyperspace":
		writeState(noErr(func() {
			chg |= cmdr.Loc.SetRef(JTarget)
			chg |= cmdr.Loc.SetRefNm(evt.MStr("StarSystem"))
			chg |= cmdr.Loc.SetMode(Jump)
		}))
	case "Supercruise":
		writeState(noErr(func() {
			chg |= cmdr.Loc.SetMode(Cruise)
		}))
	}
	return chg
}

func jeSupercruiseExit(t time.Time, evt ggja.Obj) Change {
	var chg Change
	writeState(noErr(func() {
		if bdy := evt.Str("Body", ""); bdy == "" {
			chg |= cmdr.Loc.SetRef(Space)
			chg |= cmdr.Loc.SetRefNm("?")
		} else {
			bty := evt.MStr("BodyType")
			chg |= cmdr.Loc.SetRef(jeBodyTypeMap[bty])
			chg |= cmdr.Loc.SetRefNm(bdy)
		}
	}))
	return chg
}

func jeLoadGame(t time.Time, evt ggja.Obj) Change {
	fid := evt.MStr("FID")
	nm := evt.MStr("Commander")
	writeState(noErr(func() {
		if fid != cmdr.Fid {
			cmdr.switchTo(fid, nm)
		}
		cmdr.Ship.Ship = ship.TheShips.Load(evt.MInt("ShipID"), evt.MStr("Ship"))
	}))
	return ChgCmdr | ChgShip
}

func jeLeaveBody(t time.Time, evt ggja.Obj) Change {
	var chg Change
	writeState(noErr(func() {
		chg |= cmdr.Loc.SetRef(Space)
		chg |= cmdr.Loc.SetRefNm("")
	}))
	return chg
}

func jeApproachSettlement(t time.Time, evt ggja.Obj) Change {
	var chg Change
	writeState(noErr(func() {
		if watched.FlagsAny(cmdr.statFlags, watched.StatFlagSupercruise) {
			return
		}
		chg |= cmdr.Loc.SetRef(Settlement)
		chg |= cmdr.Loc.SetRefNm(evt.MStr("Name"))
	}))
	return chg
}

func jeJoinACrew(t time.Time, evt ggja.Obj) Change {
	var chg Change
	writeState(noErr(func() {
		chg |= cmdr.Loc.SetVehicle(AsCrew)
	}))
	return chg
}

func jeEndCrewSession(t time.Time, evt ggja.Obj) Change {
	var chg Change
	writeState(noErr(func() {
		chg |= cmdr.Loc.SetVehicle(InShip)
	}))
	return chg
}

func cmdExpand(arr ggja.GenArr, vars bool, evt ggja.Obj) string {
	var sb strings.Builder
	for _, tmp := range arr {
		switch e := tmp.(type) {
		case string:
			switch e {
			case "dir":
				sb.WriteString(filepath.Dir(evt.MStr("Filename")))
			case "file":
				sb.WriteString(filepath.Base(evt.MStr("Filename")))
			case "basename":
				fnm := filepath.Base(evt.MStr("Filename"))
				ext := filepath.Ext(fnm)
				if ext != "" {
					fnm = fnm[:len(fnm)-len(ext)]
				}
				sb.WriteString(fnm)
			default:
				log.Panica("unknown `expansion`", e)
			}
		case ggja.GenArr:
			sb.WriteString(cmdExpand(e, !vars, evt))
		default:
			log.Panica("cmd cannot expand `element` in array", tmp)
		}
	}
	return sb.String()
}

var (
	onScreenShots struct {
		cmd      string
		argTmpls []*template.Template
	}
	fsnameReplacer = strings.NewReplacer(
		" ", "_",
		"\t", "_",
		"/", "_",
		`\`, "_",
		":", "_",
		";", "_",
	)
	onScreenShotsFuncs = template.FuncMap{
		"dir":  filepath.Dir,
		"file": filepath.Base,
		"base": func(p string) string {
			p = filepath.Base(p)
			if ext := filepath.Ext(p); ext != "" {
				p = p[:len(p)-len(ext)]
			}
			return p
		},
		"ext": filepath.Ext,
		"tshort": func(t string) (string, error) {
			ts, err := time.Parse("2006-01-02T15:04:05Z", t)
			if err != nil {
				return "", err
			}
			return ts.Format("20060102150405"), nil
		},
		"fsname": fsnameReplacer.Replace,
	}
)

func initOnScreenshots(cmd ggja.GenArr) {
	onScreenShots.cmd = cmd[0].(string)
	onScreenShots.argTmpls = make([]*template.Template, len(cmd)-1)
	var err error
	for i, t := range cmd[1:] {
		tstr := t.(string)
		log.Debuga("parse onScreenshot `arg#` `from`", i, tstr)
		nm := fmt.Sprintf("arg%d", i)
		tmpl := template.New(nm).Funcs(onScreenShotsFuncs)
		tmpl, err = tmpl.Parse(tstr)
		if err != nil {
			log.Errore(err)
		}
		onScreenShots.argTmpls[i] = tmpl
	}
}

func jeScreenshot(t time.Time, evt ggja.Obj) Change {
	if onScreenShots.argTmpls == nil {
		var scCmd ggja.GenArr
		readState(noErr(func() { scCmd = cmdr.OnScreenShot }))
		if len(scCmd) == 0 {
			onScreenShots.argTmpls = []*template.Template{}
			return 0
		}
		initOnScreenshots(scCmd)
	}
	if onScreenShots.cmd == "" {
		return 0
	}
	args := make([]string, len(onScreenShots.argTmpls))
	for i := range args {
		var sb strings.Builder
		err := onScreenShots.argTmpls[i].Execute(&sb, evt.Bare)
		if err != nil {
			log.Panice(err)
		}
		args[i] = sb.String()
	}
	log.Debuga("exec `cmd` with `args`", onScreenShots.cmd, args)
	go func() {
		exe := exec.Command(onScreenShots.cmd, args...)
		err := exe.Run()
		if err != nil {
			log.Errore(err)
		}
	}()
	return 0
}

var (
	jeBodyTypeMap = map[string]PosRef{
		"Star":            Star,
		"Planet":          Planet,
		"PlanetaryRing":   Ring,
		"StellarRing":     Ring,
		"Station":         Station,
		"AsteroidCluster": Belt,
	}
	jeStnTypeMap = map[string]PosRef{
		"Coriolis":      Station,
		"Orbis":         Station,
		"Outpost":       Outpost,
		"CraterOutpost": Outpost,
		"CraterPort":    Station,
	}
)
