package main

import (
	"os"
	"sync"

	c "github.com/CmdrVasquess/BCplus/cmdr"
	l "github.com/fractalqb/qblog"
	robi "github.com/go-vgo/robotgo"
)

type userHanlder func(*c.GmState, map[string]interface{}) (reload bool)

func DispatchUser(lock *sync.RWMutex, state *c.GmState, event map[string]interface{}) {
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
		case "all":
			handler, _ = allUsrOps[oprtn]
		case "travel":
			handler, _ = trvlUsrOps[oprtn]
		case "materials":
			handler, _ = matUsrOps[oprtn]
		case "synth":
			handler, _ = synUsrOps[oprtn]
		}
	}
	if handler == nil {
		eulog.Logf(l.Warn, "no handler for user event: topic=%v; op=%s", topic, oprtn)
	} else {
		eulog.Log(l.Debug, "handling event:", event)
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

var allUsrOps = map[string]userHanlder{
	"skbd":    allSkbd,
	"tglhome": allTglHome,
	"tgldest": allTglDest,
	"quit":    allQuit,
}

const (
	tagHome  = "Home"
	tagAbndn = "abandoned"
)

func allTglHome(gstat *c.GmState, evt map[string]interface{}) (reload bool) {
	cmdr := &gstat.Cmdr
	if cmdr.Home.Nil() {
		cmdr.Home = cmdr.Loc
		dest := cmdr.GetDest(cmdr.Home.Ref)
		dest.Tag(tagHome)
		dest.Untag(tagAbndn)
		reload = true
	} else if cmdr.Loc != cmdr.Home {
		dest := cmdr.GetDest(cmdr.Home.Ref)
		dest.Tag(tagHome, tagAbndn)
		cmdr.Home = cmdr.Loc
		dest = cmdr.GetDest(cmdr.Home.Ref)
		dest.Tag(tagHome)
		dest.Untag(tagAbndn)
		reload = true
	} else if cmdr.Loc == cmdr.Home {
		dest := cmdr.GetDest(cmdr.Home.Ref)
		dest.Tag(tagHome, tagAbndn)
		cmdr.Home.Ref = nil
		reload = true
	}
	return reload
}

func allTglDest(gstat *c.GmState, evt map[string]interface{}) (reload bool) {
	cmdr := &gstat.Cmdr
	dest := cmdr.FindDest(cmdr.Loc.Ref)
	if dest == nil {
		dest = cmdr.GetDest(cmdr.Loc.Ref)
		if !cmdr.Home.Nil() && dest.Loc.String() == cmdr.Home.Ref.String() {
			dest.Tag(tagHome)
		}
	} else {
		cmdr.RmDest(cmdr.Loc.Ref)
	}
	return true
}

func allSkbd(gstat *c.GmState, evt map[string]interface{}) (reload bool) {
	txt, _ := attStr(evt, "str")
	eulog.Logf(l.Trace, "sending as keyboard input: [%s]", txt)
	robi.TypeStr(txt)
	return false
}

func allQuit(gstat *c.GmState, evt map[string]interface{}) (reload bool) {
	myPid := os.Getpid()
	me, _ := os.FindProcess(myPid)
	eulog.Logf(l.Debug, "user quit: sending signal %s to %d", os.Interrupt, myPid)
	me.Signal(os.Interrupt)
	return false
}
