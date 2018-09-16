package main

import (
	"git.fractalqb.de/fractalqb/ggja"
	"github.com/CmdrVasquess/BCplus/cmdr"
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
