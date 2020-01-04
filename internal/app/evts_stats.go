package app

import (
	"bytes"
	"encoding/json"
	"io/ioutil"

	"github.com/CmdrVasquess/watched"

	"git.fractalqb.de/fractalqb/ggja"
	"git.fractalqb.de/fractalqb/qbsllm"
)

var (
	selog    = qbsllm.New(qbsllm.Lnormal, "e-status", nil, nil)
	selogCfg = qbsllm.Config(selog)
)

func readStats(file string) (res ggja.GenObj) {
	evt, err := ioutil.ReadFile(file)
	if err != nil {
		log.Panice(err)
	}
	selog.Debug(bytes.NewBuffer(evt))
	err = json.Unmarshal(evt, &res)
	if err != nil {
		log.Panice(err)
	}
	return res
}

func statusEvent(statFile string) (chg Change) {
	const flagsParked = watched.StatFlagDocked | watched.StatFlagLanded
	defer recoverEvent("stats", statFile)
	stats := ggja.Obj{Bare: readStats(statFile)}
	writeState(noErr(func() {
		flags := stats.MUint32("Flags")
		cmdr.statFlags = flags
		switch {
		case watched.FlagsAny(flags, watched.StatFlagSupercruise):
			chg |= cmdr.Head.Loc.SetMode(Cruise)
		case watched.FlagsAny(flags, flagsParked):
			chg |= cmdr.Head.Loc.SetMode(Parked)
		default:
			chg |= cmdr.Head.Loc.SetMode(Move)
		}
		switch {
		case watched.FlagsAny(flags, watched.StatFlagInMainShip):
			chg |= cmdr.Head.Loc.SetVehicle(InShip)
		case watched.FlagsAny(flags, watched.StatFlagInSrv):
			chg |= cmdr.Head.Loc.SetVehicle(InSRV)
		case watched.FlagsAny(flags, watched.StatFlagInFighter):
			chg |= cmdr.Head.Loc.SetVehicle(InFighter)
		default:
			chg |= cmdr.Head.Loc.SetVehicle(AsCrew)
		}
		if watched.FlagsAny(flags, watched.StatFlagHasLatLon) {
			chg |= cmdr.surfLoc.SetAlt(stats.MF64("Altitude"))
			chg |= cmdr.surfLoc.SetLatLon(
				stats.MF64("Latitude"),
				stats.MF64("Longitude"),
			)
			if watched.FlagsAny(flags, watched.StatFlagDocked) {
				chg |= cmdr.Head.Loc.SetSurf(nil)
			} else {
				chg |= cmdr.Head.Loc.SetSurf(&cmdr.surfLoc)
			}
		} else {
			chg |= cmdr.Head.Loc.SetSurf(nil)
		}
	}))
	return chg
}
