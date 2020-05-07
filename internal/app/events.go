package app

import (
	"git.fractalqb.de/fractalqb/c4hgol"
	"github.com/CmdrVasquess/bcplus/itf"
	"github.com/CmdrVasquess/watched"
)

type EventSrc rune

const (
	ESRC_JOURNAL EventSrc = watched.EscrJournal
	ESRC_JSTATUS EventSrc = watched.EscrStatus
	ESRC_WEBUI            = 'u'
)

type Event struct {
	Src  EventSrc
	Data interface{}
}

type Change uint32

const (
	ChgCmdr Change = (1 << iota)
	ChgShip
	ChgCoos
	ChgLoc
	WuiUpInSys
	WuiUpTrvl
)

var (
	EventQ  = make(chan Event, 16)
	elogCfg = c4hgol.NewGroup("events", jelogCfg, selogCfg)
)

func watchJournalDir(dir string) (quit chan<- bool) {
	res := &watched.JournalDir{
		Dir: dir,
		PerJLine: func(line []byte) {
			EventQ <- Event{Src: ESRC_JOURNAL, Data: string(line)}
		},
		OnStatChg: func(tag rune, statFile string) {
			EventQ <- Event{Src: EventSrc(tag), Data: statFile}
		},
		Quit: make(chan bool),
	}
	last, err := watched.NewestJournal(dir)
	if err != nil {
		log.Fatale(err)
	}
	go res.Watch(last)
	return res.Quit
}

func eventLoop() {
	log.Debugs("running main event loop")
	var (
		chg   Change
		useTs = true
	)
	for e := range EventQ {
		switch e.Src {
		case ESRC_JOURNAL:
			data := e.Data.(string)
			chg, useTs = journalEvent(data, useTs)
		case ESRC_JSTATUS:
			chg = statusEvent(e.Data.(string))
		case ESRC_WEBUI:
			webuiEvent(e.Data.([]byte))
		default:
			log.Debuga("drop event from `source`", string(e.Src))
		}
		if chg != 0 {
			webUiUpd <- chg
			if chg&ChgLoc == ChgLoc {
				edpcStub.LocUpdate <- itf.Location{
					SysID:    cmdr.Loc.Sys.Addr,
					SysName:  cmdr.Loc.Sys.Name,
					LocInSys: cmdr.Loc.LocInSys,
				}
			}
		}
		chg = 0
	}
}

func recoverEvent(source string, hint interface{}) {
	if r := recover(); r != nil {
		if hint == nil {
			log.Errora("panic handling `src`: `err`", source, r)
		} else {
			log.Errora("panic handling `src`: `err` (`hint`)", source, r, hint)
		}
	}
}
