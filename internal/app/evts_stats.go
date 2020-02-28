package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"time"

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

var lastSurfaceHint time.Time

func speakSurfaceHint(stats ggja.Obj) {
	if len(cmdr.SurfDest) < 2 {
		return
	}
	now := time.Now()
	if !lastSurfaceHint.IsZero() && now.Sub(lastSurfaceHint) < 10*time.Second {
		return
	}
	bear := BearingDeg(
		cmdr.surfLoc.LatLon[0], cmdr.surfLoc.LatLon[1],
		cmdr.SurfDest[0], cmdr.SurfDest[1],
	)
	hdng := stats.MF64("Heading")
	if math.Abs(hdng-bear) > 3 {
		lastSurfaceHint = now
		dispatchVoice(now, ChanSurf, 0,
			fmt.Sprintf("Hold bearing: %d°", int(bear)))
	}
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
			chg |= cmdr.Loc.SetMode(Cruise)
		case watched.FlagsAny(flags, flagsParked):
			chg |= cmdr.Loc.SetMode(Parked)
		default:
			chg |= cmdr.Loc.SetMode(Move)
		}
		switch {
		case watched.FlagsAny(flags, watched.StatFlagInMainShip):
			chg |= cmdr.Loc.SetVehicle(InShip)
		case watched.FlagsAny(flags, watched.StatFlagInSrv):
			chg |= cmdr.Loc.SetVehicle(InSRV)
		case watched.FlagsAny(flags, watched.StatFlagInFighter):
			chg |= cmdr.Loc.SetVehicle(InFighter)
		default:
			chg |= cmdr.Loc.SetVehicle(AsCrew)
		}
		if watched.FlagsAny(flags, watched.StatFlagHasLatLon) {
			chg |= cmdr.surfLoc.SetAlt(stats.MF64("Altitude"))
			chg |= cmdr.surfLoc.SetLatLon(
				stats.MF64("Latitude"),
				stats.MF64("Longitude"),
			)
			if watched.FlagsAny(flags, watched.StatFlagDocked) {
				chg |= cmdr.Loc.SetSurf(nil)
			} else {
				chg |= cmdr.Loc.SetSurf(&cmdr.surfLoc)
				speakSurfaceHint(stats)
			}
		} else {
			chg |= cmdr.Loc.SetSurf(nil)
		}
	}))
	return chg
}
