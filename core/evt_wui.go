package core

import (
	"git.fractalqb.de/fractalqb/ggja"
	"github.com/CmdrVasquess/BCplus/cmdr"
	"github.com/CmdrVasquess/BCplus/galaxy"
	robi "github.com/go-vgo/robotgo"
)

func userEvent(jEvt ggja.GenObj) {
	evt := ggja.Obj{Bare: jEvt}
	switch cmd := evt.MStr("cmd"); cmd {
	case "kbd":
		txt := evt.MStr("txt")
		log.Debugf("send as keyboard input: '%s'", txt)
		robi.TypeStr(txt)
	case "synth-demand":
		userSynthDemand(evt)
	case "surf-dest":
		userSurfaceDest(evt)
	case "switch-home":
		userSwitchHome(evt)
	default:
		log.Errorf("unknown user input command '%s'", cmd)
	}
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

func userSurfaceDest(evt ggja.Obj) {
	stateLock.Lock()
	defer stateLock.Unlock()
	theCmdr.Surface.Dest = [2]cmdr.CooNaN{
		cmdr.CooNaN(evt.F32("lat", galaxy.NaN32)),
		cmdr.CooNaN(evt.F32("lon", galaxy.NaN32)),
	}
	theCmdr.Surface.Box = evt.F32("box", 0)
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
