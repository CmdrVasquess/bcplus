package main

import (
	"bufio"
	"os"
	"reflect"
	"time"

	l "github.com/fractalqb/qblog"
	"github.com/fractalqb/xsx"
	"github.com/fractalqb/xsx/gem"
	"github.com/fractalqb/xsx/table"
	robi "github.com/go-vgo/robotgo"
)

var jMacros = make(map[string]*gem.Sequence)
var macroPause time.Duration = 100

func loadMacros(defFileName string) {
	def, err := os.Open(defFileName)
	if err != nil {
		glog.Logf(l.Warn, "cannot read macros: %s", err.Error())
		return
	}
	defer def.Close()
	xpp := xsx.NewPullParser(bufio.NewReader(def))
	tDef, err := table.ReadDef(xpp)
	if err != nil {
		glog.Logf(l.Error, "macro file: %s", err.Error())
		return
	}
	for row, err := tDef.NextRow(xpp, nil); row != nil; row, err = tDef.NextRow(xpp, row) {
		if err != nil {
			glog.Logf(l.Error, "macro row: %s", err.Error())
			return
		}
		switch row[0].(*gem.Atom).Str {
		case "j":
			evtNm := row[1].(*gem.Atom).Str
			macro := row[2].(*gem.Sequence)
			jMacros[evtNm] = macro
		default:
			glog.Logf(l.Warn, "unsupported source for macro event: '%s'",
				row[0].(*gem.Atom).Str)
		}
	}
	glog.Logf(l.Info, "%d journal macros loaded", len(jMacros))
}

func playMacro(m *gem.Sequence, hint string) {
	for _, step := range m.Elems {
		switch s := step.(type) {
		case *gem.Atom:
			if s.Quoted() {
				glog.Logf(l.Trace, "macro '%s' type string \"%s\"", hint, s.Str)
				robi.TypeStr(s.Str)
			} else {
				glog.Logf(l.Trace, "macro '%s' tab key %s", hint, s.Str)
				robi.KeyTap(s.Str)
			}
		case *gem.Sequence:
			if s.Meta() {
				glog.Logf(l.Warn, "macro  '%s' has meta sequence", hint)
			} else {
				switch s.Brace() {
				case '{':
					playMouse(s, hint)
				case '[':
					play2Proc(s, hint)
				}
			}
		default:
			glog.Logf(l.Warn, "macro '%s': unhandled element type: %s",
				hint,
				reflect.TypeOf(step))
		}
		time.Sleep(macroPause * time.Millisecond) // TODO make it adjustable
	}
}

func playMouse(m *gem.Sequence, hint string) {
	//	for _, step := range m.Elems {
	//		actn := step.(*gem.Atom).Str

	//	}
}

func play2Proc(s *gem.Sequence, hint string) {
	if len(s.Elems) > 0 {
		// TODO: switching seems to not yet work?
		procNm := s.Elems[0].(*gem.Atom).Str
		glog.Logf(l.Debug, "macro switch to process '%s'", procNm)
		current := robi.GetActive()
		robi.ActiveName(procNm)
		defer func() {
			glog.Logf(l.Debug, "macro switch back from '%s'", procNm)
			robi.SetActive(current)
		}()
		rest := gem.Sequence{}
		rest.Elems = s.Elems[1:]
		playMacro(&rest, hint)
	}
}

func jEventMacro(evtName string) {
	macro, ok := jMacros[evtName]
	if ok {
		glog.Logf(l.Debug, "play journal event macro: %s", evtName)
		go playMacro(macro, evtName)
	}
}
