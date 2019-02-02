package core

import (
	"sort"

	"git.fractalqb.de/fractalqb/ggja"
	"github.com/CmdrVasquess/BCplus/cmdr"
	"github.com/CmdrVasquess/BCplus/galaxy"
	robi "github.com/go-vgo/robotgo"
)

var wuiEventHdl = map[string]func(ggja.Obj){
	"kbd":              userSend2Kbd,
	"synth-demand":     userSynthDemand,
	"surf-dest":        userSurfDest,
	"surf-dest-del":    userSurfDestDel,
	"surf-dest-up":     userSurfDestUp,
	"surf-dest-down":   userSurfDestDown,
	"surf-dest-switch": userSurfSwitch,
	"surf-goal":        userSurfaceGoal,
	"switch-home":      userSwitchHome,
}

func userEvent(jEvt ggja.GenObj) {
	evt := ggja.Obj{Bare: jEvt}
	cmd := wuiEventHdl[evt.MStr("cmd")]
	if cmd == nil {
		lgr.Errora("unknown user input `command`", cmd)
	} else {
		cmd(evt)
	}
}

func userSend2Kbd(evt ggja.Obj) {
	txt := evt.MStr("txt")
	lgr.Debuga("send as keyboard `input`", txt)
	robi.TypeStr(txt)
}

func userSynthDemand(evt ggja.Obj) {
	stateLock.Lock()
	defer stateLock.Unlock()
	recipe := cmdr.DefRecipe(cmdr.RcpSynth, evt.MStr("recipe"))
	var dmnd []int
	for _, jn := range evt.MArr("demand").Bare {
		dmnd = append(dmnd, int(jn.(float64)))
	}
	theCmdr.RcpDmnd[recipe] = dmnd
}

func userSurfaceGoal(evt ggja.Obj) {
	theCmdr.Surface.Goal = evt.MInt("idx")
	if theCmdr.Surface.Goal < 0 {
		theCmdr.Surface.Goal = 0
	} else if theCmdr.Surface.Goal >= len(theCmdr.Surface.Dests) {
		theCmdr.Surface.Goal = len(theCmdr.Surface.Dests) - 1
	}
}

func userSurfDest(evt ggja.Obj) {
	stateLock.Lock()
	defer stateLock.Unlock()
	point := cmdr.SurfDest{
		Tag: evt.MStr("tag"),
		Point: [2]cmdr.CooNaN{
			cmdr.CooNaN(evt.F32("lat", galaxy.NaN32)),
			cmdr.CooNaN(evt.F32("lon", galaxy.NaN32)),
		},
		Box: evt.F32("box", 0),
	}
	if idx := evt.Int("idx", -1); idx >= len(theCmdr.Surface.Dests) || idx < 0 {
		theCmdr.Surface.Dests = append(theCmdr.Surface.Dests, point)
	} else {
		theCmdr.Surface.Dests[idx] = point
	}
}

func userSurfSwitch(evt ggja.Obj) {
	stateLock.Lock()
	defer stateLock.Unlock()
	theCmdr.Surface.Switch = evt.MBool("flag")
}

func userSurfDestDel(evt ggja.Obj) {
	stateLock.Lock()
	defer stateLock.Unlock()
	delArr := evt.MArr("idxs")
	if delArr.Len() == 0 {
		return
	}
	dels := make([]int, delArr.Len())
	for i := range dels {
		dels[i] = delArr.MInt(i)
	}
	sort.Ints(dels)
	di := 0
	var newDets []cmdr.SurfDest
	for i, dst := range theCmdr.Surface.Dests {
		if di >= len(dels) || dels[di] != i {
			newDets = append(newDets, dst)
		} else {
			di++
		}
	}
	theCmdr.Surface.Dests = newDets
}

func userSurfDestUp(evt ggja.Obj) {
	stateLock.Lock()
	defer stateLock.Unlock()
	idx := evt.MInt("idx")
	if idx > 0 {
		ds := theCmdr.Surface.Dests
		ds[idx-1], ds[idx] = ds[idx], ds[idx-1]
		theCmdr.Surface.Goal = evt.MInt("goal")
	}
}

func userSurfDestDown(evt ggja.Obj) {
	stateLock.Lock()
	defer stateLock.Unlock()
	idx := evt.MInt("idx")
	ds := theCmdr.Surface.Dests
	if idx+1 < len(ds) {
		ds[idx], ds[idx+1] = ds[idx+1], ds[idx]
		theCmdr.Surface.Goal = evt.MInt("goal")
	}
}

func userSwitchHome(evt ggja.Obj) {
	stateLock.Lock()
	defer stateLock.Unlock()
	home := evt.Str("home", "")
	if len(home) == 0 {
		theCmdr.Home.Clear()
	} else {
		theCmdr.Home = theCmdr.Loc
	}
}
