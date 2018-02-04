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
	if err := gxw.ParseHtmlTemplate(assetPath("tk.html"), "tk", tmpls); err != nil {
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
		bt = gxTkLocSys.NewBounT()
		bt.BindP(gxTkLocSys.Name, gxw.HtmlEsc(loc.Name()))
	case *gxy.SysBody:
		bt = gxTkLocBdy.NewBounT()
		bt.BindP(gxTkLocBdy.Sys, gxw.HtmlEsc(loc.System().Name()))
		bt.BindP(gxTkLocBdy.Name, gxw.HtmlEsc(loc.Name))
	case *gxy.Station:
		bt = gxTkLocStn.NewBounT()
		bt.BindP(gxTkLocBdy.Sys, gxw.HtmlEsc(loc.System().Name()))
		bt.BindP(gxTkLocStn.Name, loc.Name)
	default:
		panic(reflect.TypeOf(l.loc).String())
	}
	n += bt.Emit(wr)
	return n
}
