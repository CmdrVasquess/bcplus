package main

import (
	"io"
	"reflect"

	gxy "github.com/CmdrVasquess/BCplus/galaxy"
	gx "github.com/fractalqb/goxic"
	gxw "github.com/fractalqb/goxic/web"
)

var gxTkLocSys struct {
	*gx.Template
	Name []int
}

var gxTkLocBdy struct {
	*gx.Template
	Sys  []int
	Name []int
}

var gxTkLocStn struct {
	*gx.Template
	Sys  []int
	Name []int
}

func loadWebTkTemplates() {
	tmpls := make(map[string]*gx.Template)
	tpars := gxw.NewHtmlParser()
	if err := tpars.ParseFile(assetPath("tk.html"), "tk", tmpls); err != nil {
		panic("failed loading templates: " + err.Error())
	}
	gx.MustIndexMap(&gxTkLocSys, needTemplate(tmpls, "loc-system"), idxMapNames.Convert)
	gx.MustIndexMap(&gxTkLocBdy, needTemplate(tmpls, "loc-body"), idxMapNames.Convert)
	gx.MustIndexMap(&gxTkLocStn, needTemplate(tmpls, "loc-station"), idxMapNames.Convert)
}

type CntLoc struct {
	loc gxy.Location
}

func (l CntLoc) Emit(wr io.Writer) (n int) {
	var bt *gx.BounT
	switch loc := l.loc.(type) {
	case *gxy.StarSys:
		bt = gxTkLocSys.NewBounT(nil)
		bt.BindP(gxTkLocSys.Name, gxw.EscHtml(loc.Name()))
	case *gxy.SysBody:
		bt = gxTkLocBdy.NewBounT(nil)
		bt.BindP(gxTkLocBdy.Sys, gxw.EscHtml(loc.System().Name()))
		bt.BindP(gxTkLocBdy.Name, gxw.EscHtml(loc.Name))
	case *gxy.Station:
		bt = gxTkLocStn.NewBounT(nil)
		bt.BindP(gxTkLocBdy.Sys, gxw.EscHtml(loc.System().Name()))
		bt.BindP(gxTkLocStn.Name, loc.Name)
	default:
		panic(reflect.TypeOf(l.loc).String())
	}
	n += bt.Emit(wr)
	return n
}
