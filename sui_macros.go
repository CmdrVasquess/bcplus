package main

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"strings"

	c "github.com/CmdrVasquess/BCplus/cmdr"
	gx "git.fractalqb.de/fractalqb/goxic"
	gxw "git.fractalqb.de/fractalqb/goxic/web"
	l "git.fractalqb.de/fractalqb/qblog"
	"git.fractalqb.de/fractalqb/xsx"
	"git.fractalqb.de/fractalqb/xsx/gem"
)

var gxtSMcPage struct {
	*gx.Template
	Disabled []int
	Events   []int
}

var gxcDisable gx.Content

var gxtSMcMacro struct {
	*gx.Template
	Event []int
	Macro []int
	Check []int
}

func loadSMcTemplates() {
	tmpls := make(map[string]*gx.Template)
	tpars := gxw.NewHtmlParser()
	if err := tpars.ParseFile(assetPath("st-macros.html"), "macros", tmpls); err != nil {
		panic("failed loading templates: " + err.Error())
	}
	//	dynShpStyles = pgLocStyleFix(tmpls)
	//	endShpScrpit = pgEndScriptFix(tmpls)
	gx.MustIndexMap(&gxtSMcPage, needTemplate(tmpls, ""), idxMapNames.Convert)
	ttmp := needTemplate(tmpls, "disable")
	if raw, ok := ttmp.Static(); ok {
		gxcDisable = gx.Data(raw)
	} else {
		glog.Fatal("disabled message of jmacro settings is not static content")
	}
	gx.MustIndexMap(&gxtSMcMacro, needTemplate(tmpls, "event"), idxMapNames.Convert)
}

func suiMacros(w http.ResponseWriter, r *http.Request) {
	btEmit := gxtSMcPage.NewBounT(nil)
	if enableJMacros {
		btEmit.Bind(gxtSMcPage.Disabled, gx.Empty)
	} else {
		btEmit.Bind(gxtSMcPage.Disabled, gxcDisable)
	}
	btMacro := gxtSMcMacro.NewBounT(nil)
	btEmit.BindGen(gxtSMcPage.Events, func(wr io.Writer) (n int) {
		strb := bytes.NewBuffer(nil)
		xwr := xsx.Indenting(strb, " ")
		for mi := MacroName(0); mi < NO_JEVENT; mi++ {
			mnm := mi.String()
			btMacro.BindP(gxtSMcMacro.Event, mnm)
			if macro, ok := jMacros[mnm]; ok {
				strb.Reset()
				gem.Print(xwr, macro.Seq)
				mstr := strings.TrimSpace(strb.String())
				mstr = mstr[1 : len(mstr)-1]
				btMacro.BindP(gxtSMcMacro.Macro, mstr)
				if macro.Active {
					btMacro.BindP(gxtSMcMacro.Check, "checked")
				} else {
					btMacro.Bind(gxtSMcMacro.Check, gx.Empty)
				}
			} else {
				btMacro.Bind(gxtSMcMacro.Macro, gx.Empty)
				btMacro.Bind(gxtSMcMacro.Check, gx.Empty)
			}
			n += btMacro.Emit(wr)
		}
		return n
	})
	btEmit.Emit(w)
}

var stngMacros = map[string]userHanlder{
	"on":   stngMcrOn,
	"off":  stngMcrOff,
	"def":  stngMcrDef,
	"test": stngMcrTest,
}

func stngMcrOn(gstat *c.GmState, evt map[string]interface{}) (reload bool) {
	mcrNm, _ := attStr(evt, "macro")
	if macro, ok := jMacros[mcrNm]; ok {
		eulog.Log(l.Info, evt)
		macro.Active = true
	}
	return false
}

func stngMcrOff(gstat *c.GmState, evt map[string]interface{}) (reload bool) {
	mcrNm, _ := attStr(evt, "macro")
	if macro, ok := jMacros[mcrNm]; ok {
		if macro.Seq == nil || len(macro.Seq.Elems) == 0 {
			delete(jMacros, mcrNm)
		} else {
			macro.Active = false
		}
	}
	return false
}

func defn2macro(defn string) (res *gem.Sequence) {
	tmp := bytes.NewBuffer(nil)
	tmp.WriteRune('[')
	tmp.WriteString(defn)
	tmp.WriteRune(']')
	xp := xsx.NewPullParser(bufio.NewReader(tmp))
	seq, _ := gem.ReadNext(xp)
	res = seq.(*gem.Sequence)
	return res
}

// TODO error feddback to user
func stngMcrDef(gstat *c.GmState, evt map[string]interface{}) (reload bool) {
	eulog.Log(l.Info, evt)
	mcrNm, _ := attStr(evt, "macro")
	mcrDef, _ := attStr(evt, "defn")
	macro, ok := jMacros[mcrNm]
	if !ok {
		macro = &Macro{}
		jMacros[mcrNm] = macro
	}
	macro.Seq = defn2macro(mcrDef)
	return false
}

func stngMcrTest(gstat *c.GmState, evt map[string]interface{}) (reload bool) {
	if !enableJMacros {
		return false
	}
	hint, _ := attStr(evt, "macro")
	mcrDef, _ := attStr(evt, "defn")
	macro := defn2macro(mcrDef)
	eulog.Logf(l.Info, "test macro '%s': [%s]", hint, mcrDef)
	playMacro(macro, hint)
	return false
}
