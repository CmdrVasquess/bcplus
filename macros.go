package main

// Macro Syntax is based on XSX:
//
// Macro is an XSX seqeunce with elements:
// - Quote atom: Type as string
// - Unquoted atom: tab key
// - ([up|down|tap] <robotgo-key> [<robodgo-mods>...]): play key
//   https://github.com/go-vgo/robotgo/blob/master/docs/keys.md
// - {<robotgo-mouse>}: play mouse
// - [<tagret> <macro>]: play 2 proc

import (
	"bufio"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	l "git.fractalqb.de/fractalqb/qblog"
	"git.fractalqb.de/fractalqb/xsx"
	"git.fractalqb.de/fractalqb/xsx/gem"
	"git.fractalqb.de/fractalqb/xsx/table"
	robi "github.com/go-vgo/robotgo"
)

func init() {
	log.Logf(l.Info, "set keyboad delay to 50ms")
	robi.SetKeyDelay(50)
}

//go:generate stringer -type=MacroName
type MacroName uint

const (
	AfmuRepairs MacroName = iota
	ApproachSettlement
	Bounty
	BuyAmmo
	BuyDrones
	Cargo
	ChangeCrewRole
	CollectCargo
	CommitCrime
	CommunityGoal
	CommunityGoalJoin
	CommunityGoalReward
	CrewAssign
	CrewFire
	CrewHire
	CrewLaunchFighter
	CrewMemberJoins
	CrewMemberQuits
	CrewMemberRoleChange
	DatalinkScan
	DatalinkVoucher
	DataScanned
	Died
	Docked
	DockFighter
	DockingCancelled
	DockingDenied
	DockingGranted
	DockingRequested
	DockSRV
	EjectCargo
	EndCrewSession
	EngineerApply
	EngineerContribution
	EngineerCraft
	EngineerProgress
	EscapeInterdiction
	FactionKillBond
	FetchRemoteModule
	Fileheader
	Friends
	FSDJump
	FuelScoop
	HeatDamage
	HeatWarning
	HullDamage
	Interdicted
	Interdiction
	JetConeBoost
	JoinACrew
	LaunchFighter
	LaunchSRV
	Liftoff
	LoadGame
	Loadout
	Location
	MarketBuy
	MarketSell
	MassModuleStore
	MaterialCollected
	MaterialDiscarded
	MaterialDiscovered
	Materials
	MiningRefined
	MissionAbandoned
	MissionAccepted
	MissionCompleted
	MissionFailed
	MissionRedirected
	ModuleBuy
	ModuleRetrieve
	ModuleSell
	ModuleSellRemote
	ModuleStore
	ModuleSwap
	Music
	NavBeaconScan
	Passengers
	PayFines
	PayLegacyFines
	PowerplayCollect
	PowerplayDeliver
	PowerplayFastTrack
	PowerplayLeave
	PowerplaySalary
	Progress
	Promotion
	QuitACrew
	Rank
	RebootRepair
	ReceiveText
	RedeemVoucher
	RefuelAll
	Repair
	RepairAll
	RepairDrone
	RestockVehicle
	Resurrect
	Scan
	Scanned
	Screenshot
	SearchAndRescue
	SellDrones
	SellExplorationData
	SendText
	SetUserShipName
	ShieldState
	ShipyardBuy
	ShipyardNew
	ShipyardSell
	ShipyardSwap
	ShipyardTransfer
	StartJump
	SupercruiseEntry
	SupercruiseExit
	Synthesis
	Touchdown
	Undocked
	USSDrop
	VehicleSwitch
	WingAdd
	WingInvite
	WingJoin
	WingLeave
	NO_JEVENT
)

type Macro struct {
	fMask, fVal uint32
	seq         *gem.Sequence
}

var jMacros = make(map[string]*Macro)
var macroPause = 50 * time.Millisecond

const (
	MCR_COLNM_SRC  = "source"
	MCR_COLNM_EVT  = "event"
	MCR_COLNM_FMSK = "flags-mask"
	MCR_COLNM_FVAL = "flags-value"
	MCR_COLNM_MCR  = "macro"
)

// TODO error handling
func saveMacros(toFileName string) {
	log.Logf(l.Info, "save macros to %s", toFileName)
	tmpfn := toFileName + "~"
	wr, err := os.Create(tmpfn)
	if err != nil {
		log.Logf(l.Error, "failed to save macros to '%s'", tmpfn)
		return
	}
	defer func() {
		if wr != nil {
			wr.Close()
		}
	}()
	xwr := xsx.Indenting(wr, "\t")
	xsx.Print(xwr,
		xsx.B('['),
		MCR_COLNM_SRC, MCR_COLNM_EVT, MCR_COLNM_FMSK, MCR_COLNM_FVAL, MCR_COLNM_MCR,
		xsx.End)
	xwr.Newline(1, 0)
	evtSort := make([]string, 0, len(jMacros))
	for e := range jMacros {
		evtSort = append(evtSort, e)
	}
	sort.Strings(evtSort)
	for _, eNm := range evtSort {
		e := jMacros[eNm]
		xwr.Begin('(', false)
		xwr.Atom("j", false, xsx.Qcond)
		xwr.Atom(eNm, false, xsx.Qcond)
		xwr.Atom(strconv.FormatUint(uint64(e.fMask), 16), false, xsx.QSUPPRESS)
		xwr.Atom(strconv.FormatUint(uint64(e.fVal), 16), false, xsx.QSUPPRESS)
		gem.Print(xwr, e.seq)
		xwr.End()
		xwr.Newline(1, 0)
	}
	wr.Close()
	wr = nil
	os.Rename(tmpfn, toFileName)
}

func loadMacros(defFileName string) {
	def, err := os.Open(defFileName)
	if err != nil {
		log.Logf(l.Warn, "cannot read macros: %s", err.Error())
		return
	}
	defer def.Close()
	xpp := xsx.NewPullParser(bufio.NewReader(def))
	tDef, err := table.ReadDef(xpp)
	if err != nil {
		log.Logf(l.Error, "macro file: %s", err.Error())
		return
	}
	colSrc := tDef.ColIndex(MCR_COLNM_SRC)
	if colSrc < 0 {
		log.Log(l.Error, "macro definition has no column 'source'")
		return
	}
	colEvt := tDef.ColIndex(MCR_COLNM_EVT)
	if colEvt < 0 {
		log.Log(l.Error, "macro definition has no column 'event'")
		return
	}
	colFMask := tDef.ColIndex(MCR_COLNM_FMSK)
	if colFMask < 0 {
		log.Log(l.Error, "macro definition has no column 'flags-mask'")
		return
	}
	colFVal := tDef.ColIndex(MCR_COLNM_FVAL)
	if colFVal < 0 {
		log.Log(l.Error, "macro definition has no column 'flags-value'")
		return
	}
	colMcr := tDef.ColIndex(MCR_COLNM_MCR)
	if colMcr < 0 {
		log.Log(l.Error, "macro definition has no column 'macro'")
		return
	}
	actvNo := 0
	for row, err := tDef.NextRow(xpp, nil); row != nil; row, err = tDef.NextRow(xpp, row) {
		if err != nil {
			log.Logf(l.Error, "macro row: %s", err.Error())
			return
		}
		switch row[colSrc].(*gem.Atom).Str {
		case "j":
			evtNm := row[colEvt].(*gem.Atom).Str
			fMask, err := strconv.ParseUint(row[colFMask].(*gem.Atom).Str, 16, 32)
			if err != nil {
				log.Logf(l.Error, "invalid state flags mask: %s", err.Error())
				return
			}
			fVal, err := strconv.ParseUint(row[colFVal].(*gem.Atom).Str, 16, 32)
			if err != nil {
				log.Logf(l.Error, "invalid state flags value: %s", err.Error())
				return
			}
			macro := &Macro{
				fMask: uint32(fMask),
				fVal:  uint32(fVal),
				seq:   row[colMcr].(*gem.Sequence),
			}
			if (^fMask & fVal) == 0 {
				actvNo++
			}
			jMacros[evtNm] = macro
		default:
			log.Logf(l.Warn, "unsupported source for macro event: '%s'",
				row[0].(*gem.Atom).Str)
		}
	}
	log.Logf(l.Info, "%d journal macros loaded, %d active", len(jMacros), actvNo)
}

func playMacro(m *gem.Sequence, hint string) {
	for _, step := range m.Elems {
		switch s := step.(type) {
		case *gem.Atom:
			if s.Quoted() {
				log.Logf(l.Trace, "macro '%s' type string \"%s\"", hint, s.Str)
				robi.TypeStr(s.Str)
			} else {
				log.Logf(l.Trace, "macro '%s' tab key %s", hint, s.Str)
				robi.KeyTap(s.Str)
			}
		case *gem.Sequence:
			if s.Meta() {
				log.Logf(l.Warn, "macro  '%s' has meta sequence", hint)
			} else {
				switch s.Brace() {
				case '(':
					playKey(s, hint)
				case '{':
					playMouse(s, hint)
				case '[':
					play2Proc(s, hint)
				}
			}
		default:
			log.Logf(l.Warn, "macro '%s': unhandled element type: %s",
				hint,
				reflect.TypeOf(step))
		}
		time.Sleep(macroPause) // TODO make it adjustable
	}
}

func playKey(m *gem.Sequence, hint string) {
	if len(m.Elems) == 0 {
		log.Logf(l.Warn,
			"empty key sequence in macro '%s'",
			hint)
		return
	}
	var cmd []string
	action := 0
	modsAt := 1
	e := m.Elems[0].(*gem.Atom)
	if e.Meta() {
		switch e.Str {
		case "down":
			action = -1
			cmd = append(cmd, "down")
		case "up":
			action = 1
			cmd = append(cmd, "up")
		case "tap":
			action = 0
		default:
			log.Logf(l.Error,
				"unknown key action '%s' in macro '%s'",
				e.Str,
				hint)
			return
		}
		if len(m.Elems) < 2 {
			log.Logf(l.Error,
				"missing key spec in key sequence of macro '%s'",
				hint)
		}
		cmd = append(cmd, m.Elems[1].(*gem.Atom).Str)
		modsAt = 2
	} else {
		cmd = append(cmd, e.Str)
	}
	for _, e := range m.Elems[modsAt:] {
		cmd = append(cmd, e.(*gem.Atom).Str)
	}
	switch action {
	case 0:
		robi.KeyTap(cmd[0], cmd[1:])
	default:
		tmp := cmd[0]
		cmd[0] = cmd[1]
		cmd[1] = tmp
		robi.KeyToggle(cmd...)
	}
}

func playMouse(m *gem.Sequence, hint string) {
	for ip := 0; ip < len(m.Elems); ip++ {
		switch m.Elems[ip].(*gem.Atom).Str {
		case "left":
			ip++
			mouseButton("left", m.Elems[ip].(*gem.Atom).Str)
		case "middle":
			ip++
			mouseButton("center", m.Elems[ip].(*gem.Atom).Str)
		case "right":
			ip++
			mouseButton("right", m.Elems[ip].(*gem.Atom).Str)
		case "click":
			ip++
			xk, yk := mouseCoos(
				m.Elems[ip].(*gem.Atom).Str,
				m.Elems[ip+1].(*gem.Atom).Str)
			ip++
			robi.MoveMouse(xk, yk)
		case "drag":
			ip++
			xk, yk := mouseCoos(
				m.Elems[ip].(*gem.Atom).Str,
				m.Elems[ip+1].(*gem.Atom).Str)
			ip++
			robi.DragMouse(xk, yk)
		case "scroll":
			ip++
			count, _ := strconv.ParseInt(m.Elems[ip].(*gem.Atom).Str, 10, 32)
			ip++
			dir := m.Elems[ip].(*gem.Atom).Str
			robi.ScrollMouse(int(count), dir)
		default:
			log.Logf(l.Error,
				"unknown mouse action: '%s'",
				m.Elems[ip].(*gem.Atom).Str)
		}
	}
}

func mouseCoos(xStr, yStr string) (x int, y int) {
	xpf := strings.ContainsAny(xStr, "+-")
	ypf := strings.ContainsAny(yStr, "+-")
	if xpf || ypf {
		x, y = robi.GetMousePos()
		if xpf {
			tmp, _ := strconv.ParseInt(xStr[1:], 10, 32)
			if xStr[0] == '+' {
				x += int(tmp)
			} else {
				x -= int(tmp)
			}
		} else {
			tmp, _ := strconv.ParseInt(xStr, 10, 32)
			x = int(tmp)
		}
		if ypf {
			tmp, _ := strconv.ParseInt(yStr[1:], 10, 32)
			if yStr[0] == '+' {
				y += int(tmp)
			} else {
				y -= int(tmp)
			}
		} else {
			tmp, _ := strconv.ParseInt(yStr, 10, 32)
			y = int(tmp)
		}
	} else {
		px, err := strconv.ParseInt(xStr, 10, 32)
		if err != nil {
			log.Logf(l.Error, "parse mouse x-coo '%s'", xStr)
		}
		py, err := strconv.ParseInt(yStr, 10, 32)
		if err != nil {
			log.Logf(l.Error, "parse mouse y-coo '%s'", yStr)
		}
		x = int(px)
		y = int(py)
	}
	return x, y
}

func mouseButton(which string, action string) {
	switch action {
	case "click":
		robi.MouseClick(which, false)
	case "double":
		robi.MouseClick(which, true)
	case "down":
		robi.MouseToggle("down", which)
	case "up":
		robi.MouseToggle("up", which)
	default:
		log.Logf(l.Error,
			"unknown mouse-button action: '%s'",
			action)
	}
}

func play2Proc(s *gem.Sequence, hint string) {
	if len(s.Elems) > 0 {
		// TODO: switching seems to not yet work?
		procNm := s.Elems[0].(*gem.Atom).Str
		log.Logf(l.Debug, "macro switch to process '%s'", procNm)
		current := robi.GetActive()
		robi.ActiveName(procNm)
		defer func() {
			log.Logf(l.Debug, "macro switch back from '%s'", procNm)
			robi.SetActive(current)
		}()
		rest := gem.Sequence{}
		rest.Elems = s.Elems[1:]
		playMacro(&rest, hint)
	}
}

func jEventMacro(evtName string, jFlags uint32) {
	macro, ok := jMacros[evtName]
	if ok {
		jFlags &= macro.fMask
		if jFlags == macro.fVal {
			log.Logf(l.Debug, "play journal event macro: %s", evtName)
			go playMacro(macro.seq, evtName)
		}
	}
}
