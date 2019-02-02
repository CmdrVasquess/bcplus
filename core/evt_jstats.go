package core

import (
	"encoding/json"
	"math"
	"os"

	"github.com/CmdrVasquess/BCplus/webui"

	"git.fractalqb.de/fractalqb/ggja"
	log "git.fractalqb.de/fractalqb/qbsllm"
	"github.com/CmdrVasquess/BCplus/cmdr"
	"github.com/CmdrVasquess/BCplus/galaxy"
)

// https://github.com/EDCD/EDDI/issues/371
const (
	statDocked uint32 = (1 << iota)
	statLanded
	statGearDown
	statShieldsUp
	statSupercruise

	statFAOff
	statHPDeployed
	statInWing
	statLightsOn
	statCSDeployed

	statSilentRun
	statFuelScooping
	statSrvHandbrake
	statSrvTurret
	statSrvUnderShip

	statSrvDriveAssist
	statFsdMassLock
	statFsdCharging
	statCooldown
	statLowFuel

	statOverHeat
	statHasLatLon
	statIsInDanger
	statInterdicted
	statInMainShip

	statInFighter
	statInSrv
)

type statGuiFocus int

//go:generate stringer -type statGuiFocus
const (
	statNoFocus statGuiFocus = iota
	statIntPanel
	statExtPanel
	statComPanel
	statRolePanel
	statStationSvc
	statGalaxyMap
	statSystemMap
)

func jstatRead(file string) (ggja.GenObj, error) {
	f, err := os.Open(file)
	if err != nil {
		lgr.Errora("cannot open stat `file`: `err`", file, err)
		return nil, err
	}
	defer f.Close()
	res := make(map[string]interface{})
	dec := json.NewDecoder(f)
	err = dec.Decode(&res)
	if err != nil {
		lgr.Errora("cannot parse stat `file`: `err`", file, err)
		return nil, err
	}
	return res, nil
}

func jstatMarket(statFile string) {
	stat, err := jstatRead(statFile)
	if err == nil && theEddn != nil {
		go eddnSendCommodities(theEddn, stat)
	}
}

func jstatShipyard(statFile string) {
	stat, err := jstatRead(statFile)
	if err == nil && theEddn != nil {
		go eddnSendShipyard(theEddn, stat)
	}
}

func jstatOutfitting(statFile string) {
	stat, err := jstatRead(statFile)
	if err == nil && theEddn != nil {
		go eddnSendOutfitting(theEddn, stat)
	}
}

func jstatStatus(statFile string) (wuiupd webui.UIUpdate) {
	jStat, err := jstatRead(statFile)
	if err != nil {
		return 0
	}
	if lgr.Logs(log.Ltrace) {
		jStr, _ := json.Marshal(jStat)
		lgr.Tracea("Status: `json`", string(jStr))
	}
	if theCmdr != nil {
		stateLock.Lock()
		defer stateLock.Unlock()
		stat := ggja.Obj{Bare: jStat}
		if flags := stat.MUint32("Flags"); flags != 0 {
			theCmdr.JStatFlags = flags
		}
		theCmdr.Loc.Alt = cmdr.CooNaN(stat.F32("Altitude", galaxy.NaN32))
		theCmdr.Loc.Lat = cmdr.CooNaN(stat.F32("Latitude", galaxy.NaN32))
		theCmdr.Loc.Lon = cmdr.CooNaN(stat.F32("Longitude", galaxy.NaN32))
		theCmdr.Loc.Heading = cmdr.CooNaN(stat.F32("Heading", galaxy.NaN32))
		if theCmdr.Surface.Switch && len(theCmdr.Surface.Dests) > 0 {
			g := &theCmdr.Surface.Dests[theCmdr.Surface.Goal]
			dLat := math.Abs(float64(theCmdr.Loc.Lat - g.Point[0]))
			dLon := math.Abs(float64(theCmdr.Loc.Lon - g.Point[1]))
			if box := float64(g.Box); dLat < box && dLon < box {
				ng := theCmdr.Surface.Goal + 1
				if ng >= len(theCmdr.Surface.Dests) {
					ng = 0
				}
				theCmdr.Surface.Goal = ng
			}
		}
	}
	return webui.UISurface
}
