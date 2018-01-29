package main

import (
	"io"
	"net/http"
	"sort"

	gxy "github.com/CmdrVasquess/BCplus/galaxy"
	gx "github.com/fractalqb/goxic"
	gxm "github.com/fractalqb/goxic/textmessage"
	gxw "github.com/fractalqb/goxic/web"
	l "github.com/fractalqb/qblog"
)

var _ *l.Logger = (*l.Logger)(nil)

var gxtSysFrame struct {
	*gx.Template
	Bodies []int
}

var gxtSysTerence struct {
	*gx.Template
}

var gxtSysStar struct {
	*gx.Template
	Name []int
	Dist []int
}

var gxtSysPlanet struct {
	*gx.Template
	Name []int
	Dist []int
	Land []int
	R    []int
	Mats []int
}

var gxtSysPMat0 struct {
	*gx.Template
	Name []int
	Frac []int
}

var gxtSysPMat1 struct {
	*gx.Template
	Name []int
	Frac []int
}

var gxtSysBelt struct {
	*gx.Template
	Name []int
	Dist []int
}

func loadSysTemplates() {
	tmpls := make(map[string]*gx.Template)
	if err := gxw.ParseHtmlTemplate(assetPath("system.html"), "system", tmpls); err != nil {
		panic("failed loading templates: " + err.Error())
	}
	//	dynShpStyles = pgLocStyleFix(tmpls)
	//	endShpScrpit = pgEndScriptFix(tmpls)
	gx.MustIndexMap(&gxtSysFrame, needTemplate(tmpls, "topic"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtSysTerence, needTemplate(tmpls, "topic/bdy-none"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtSysStar, needTemplate(tmpls, "topic/bdy-star"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtSysPlanet, needTemplate(tmpls, "topic/bdy-planet"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtSysPMat0, needTemplate(tmpls, "topic/bdy-planet/mat0"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtSysPMat1, needTemplate(tmpls, "topic/bdy-planet/mat1"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtSysBelt, needTemplate(tmpls, "topic/bdy-belt"), idxMapNames.Convert)
}

func emitBodies(wr io.Writer) (n int) {
	cmdr := &theGame.Cmdr
	if cmdr.Loc.Nil() || len(cmdr.Loc.System().Bodies) == 0 {
		btNone := gxtSysTerence.NewInitBounT(webGuiTBD)
		n += btNone.Emit(wr)
	} else {
		bdys := make([]*gxy.SysBody, len(cmdr.Loc.System().Bodies))
		copy(bdys, cmdr.Loc.System().Bodies)
		sort.Slice(bdys, func(i, j int) bool {
			di, dj := bdys[i].Dist, bdys[j].Dist
			return di < dj
		})
		btStar := gxtSysStar.NewBounT()
		btPlnt := gxtSysPlanet.NewBounT()
		btBelt := gxtSysBelt.NewBounT()
		btMat0 := gxtSysPMat0.NewBounT()
		btMat1 := gxtSysPMat1.NewBounT()
		for _, bdy := range bdys {
			switch bdy.Cat {
			case gxy.Star:
				btStar.BindP(gxtSysStar.Name, bdy.Name)
				btStar.Bind(gxtSysStar.Dist, gxm.Msg(wuiL7d, "%.2f", bdy.Dist))
				n += btStar.Emit(wr)
			case gxy.Planet:
				if bdy.IsBelt() {
					btBelt.BindP(gxtSysBelt.Name, bdy.Name)
					btBelt.Bind(gxtSysBelt.Dist, gxm.Msg(wuiL7d, "%.2f", bdy.Dist))
					n += btBelt.Emit(wr)
				} else {
					emitPlanet(wr, btPlnt, btMat0, btMat1, bdy)
				}
			default:
				btPlnt.BindP(gxtSysPlanet.Name, bdy.Name)
				btPlnt.Bind(gxtSysPlanet.Dist, gxm.Msg(wuiL7d, "%.2f", bdy.Dist))
				n += btPlnt.Emit(wr)
			}
		}
	}
	return n
}

func emitPlanet(wr io.Writer, bt, btM0, btM1 *gx.BounT, p *gxy.SysBody) (n int) {
	bt.BindP(gxtSysPlanet.Name, p.Name)
	bt.Bind(gxtSysPlanet.Dist, gxm.Msg(wuiL7d, "%.2f", p.Dist))
	bt.BindP(gxtSysPlanet.Land, p.Landable)
	bt.Bind(gxtSysPlanet.R, gxm.Msg(wuiL7d, "%.2f", p.Radius))
	bt.BindGen(gxtSysPlanet.Mats, func(wr io.Writer) (n int) {
		if len(p.Mats) == 0 {
			n += webGuiNOC.Emit(wr)
		} else {
			first := true
			for nm, f := range p.Mats {
				if first {
					btM0.BindP(gxtSysPMat0.Name, nm)
					btM0.Bind(gxtSysPMat0.Frac, gxm.Msg(wuiL7d, "%.2f", f))
					n += btM0.Emit(wr)
					first = false
				} else {
					btM1.BindP(gxtSysPMat1.Name, nm)
					btM1.Bind(gxtSysPMat1.Frac, gxm.Msg(wuiL7d, "%.2f", f))
					n += btM1.Emit(wr)
				}
			}
		}
		return n
	})
	n += bt.Emit(wr)
	return n
}

func wuiSys(w http.ResponseWriter, r *http.Request) {
	btEmit, btBind, hook := preparePage(gx.Empty, gx.Empty, activeTopic(r))
	btFrame := gxtSysFrame.NewBounT()
	btBind.Bind(hook, btFrame)
	btFrame.BindGen(gxtSysFrame.Bodies, emitBodies)
	btEmit.Emit(w)
}
