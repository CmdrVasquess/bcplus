package main

import (
	"sync"

	l "github.com/fractalqb/qblog"
)

type userHanlder func(*GmState, map[string]interface{}) (reload bool)

func DispatchUser(lock *sync.RWMutex, state *GmState, event map[string]interface{}) {
	topic, hasTopic := attStr(event, "topic")
	oprtn, hasOp := attStr(event, "op")
	if !hasOp {
		eulog.Log(l.Error, "user event without operation: ", event)
		return
	}
	eulog.Logf(l.Debug, "user event: topic=%v; op=%s", topic, oprtn)
	var handler userHanlder = nil
	if hasTopic {
		switch topic {
		case "travel":
			handler = travelPlanShip
		case "materials":
			handler = matsVis
		}
	}
	if handler == nil {
		eulog.Logf(l.Warn, "no handler for user event: topic=%v; op=%s", topic, oprtn)
	} else {
		eulog.Log(l.Info, "handling event:", event)
		lock.Lock()
		reload := handler(state, event)
		lock.Unlock() // TODO do we need defer here?
		if reload {
			eulog.Log(l.Debug, "reload after user-event")
			select {
			case wscSendTo <- true:
				eulog.Log(l.Debug, "sent web-socket event")
			default:
				eulog.Log(l.Debug, "no web-socket event sent – channel blocked")
			}
		} else {
			eulog.Log(l.Debug, "no reload after user-event")
		}
	}
}

func travelPlanShip(gstat *GmState, evt map[string]interface{}) (reload bool) {
	jshid, ok := evt["shipId"]
	if ok {
		shid := int(jshid.(float64))
		var ship *Ship = nil
		if shid >= 0 {
			ship = gstat.Cmdr.ShipById(shid)
			if ship == nil {
				eulog.Logf(l.Warn, "cannot find ship with id %d", shid)
			}
		}
		reload = (gstat.TrvlPlanShip.Ship != ship)
		eulog.Logf(l.Trace, "plan ship: %v → %v => %t",
			gstat.TrvlPlanShip.Ship,
			ship,
			reload)
		gstat.TrvlPlanShip.Ship = ship
	} else {
		eulog.Logf(l.Error, "missing ship id in travel/plan-ship")
	}
	return reload
}

func matsVis(gstat *GmState, evt map[string]interface{}) (reload bool) {
	cat, ok := attStr(evt, "cat")
	if !ok {
		eulog.Log(l.Error, "materials visibility changes has no category")
		return false
	}
	vis, ok := attStr(evt, "vis")
	if !ok {
		eulog.Log(l.Error, "materials visibility changes has no visibility")
		return false
	}
	gstat.MatCatHide[cat] = vis == "collapse"
	return false
}
