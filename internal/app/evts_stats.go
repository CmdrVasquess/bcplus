package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"time"

	"git.fractalqb.de/fractalqb/ggja"
	"git.fractalqb.de/fractalqb/qbsllm"
	"github.com/CmdrVasquess/bcplus/itf"
	"github.com/CmdrVasquess/watched"
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
	if len(cmdr.SurfDest) < 2 || cmdr.Loc.RefType != itf.Planet {
		return
	}
	now := time.Now()
	if !lastSurfaceHint.IsZero() && now.Sub(lastSurfaceHint) < 10*time.Second {
		return
	}
	if cmdr.Loc.Mode != itf.Fly && cmdr.Loc.Mode != itf.Drive {
		lastSurfaceHint = now
		dispatchVoice(now, ChanSurf, 0,
			fmt.Sprintf("Destination: %d, %d°",
				int(cmdr.SurfDest[0]),
				int(cmdr.SurfDest[1]),
			))
	} else {
		bear := BearingDeg(
			cmdr.Loc.Coos[0], cmdr.Loc.Coos[1],
			cmdr.SurfDest[0], cmdr.SurfDest[1],
		)
		hdng := stats.MF64("Heading")
		if math.Abs(hdng-bear) > 3 {
			lastSurfaceHint = now
			dispatchVoice(now, ChanSurf, 0,
				fmt.Sprintf("Hold bearing: %d°", int(bear)))
		}
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
			chg |= cmdr.setLocMode(itf.SCruise)
		case watched.FlagsAny(flags, watched.StatFlagDocked):
			chg |= cmdr.setLocMode(itf.Docked)
		default:
			chg |= cmdr.setLocMode(itf.NoTMode)
		}
		// switch {
		// case watched.FlagsAny(flags, watched.StatFlagInMainShip):
		// 	chg |= cmdr.Loc.SetVehicle(InShip)
		// case watched.FlagsAny(flags, watched.StatFlagInSrv):
		// 	chg |= cmdr.Loc.SetVehicle(InSRV)
		// case watched.FlagsAny(flags, watched.StatFlagInFighter):
		// 	chg |= cmdr.Loc.SetVehicle(InFighter)
		// default:
		// 	chg |= cmdr.Loc.SetVehicle(AsCrew)
		// }
		if watched.FlagsAny(flags, watched.StatFlagHasLatLon) {
			if watched.FlagsAny(flags, watched.StatFlagDocked) {
				chg |= cmdr.setLocCoos()
			} else {
				chg |= cmdr.setLocAlt(stats.MF64("Altitude"))
				chg |= cmdr.setLocLatLon(
					stats.MF64("Latitude"),
					stats.MF64("Longitude"),
				)
				speakSurfaceHint(stats)
			}
		} else {
			chg |= cmdr.setLocCoos()
		}
	}))
	return chg
}
