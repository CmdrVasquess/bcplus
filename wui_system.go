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
	Name   []int
	Dist   []int
	R      []int
	RotPrd []int
	Tilt   []int
}

var gxtSysPlanet struct {
	*gx.Template
	Name      []int
	Dist      []int
	Land      []int
	R         []int
	Materials []int
}

var gxtMatSec struct {
	*gx.Template
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

var dynSysStyles gx.Content

func loadSysTemplates() {
	tmpls := make(map[string]*gx.Template)
	if err := gxw.ParseHtmlTemplate(assetPath("system.html"), "system", tmpls); err != nil {
		panic("failed loading templates: " + err.Error())
	}
	dynSysStyles = pgLocStyleFix(tmpls)
	//	endShpScrpit = pgEndScriptFix(tmpls)
	gx.MustIndexMap(&gxtSysFrame, needTemplate(tmpls, "topic"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtSysTerence, needTemplate(tmpls, "topic/bdy-none"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtSysStar, needTemplate(tmpls, "topic/bdy-star"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtSysPlanet, needTemplate(tmpls, "topic/bdy-planet"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtMatSec, needTemplate(tmpls, "topic/bdy-planet/mat-section"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtSysPMat0, needTemplate(tmpls, "topic/bdy-planet/mat0"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtSysPMat1, needTemplate(tmpls, "topic/bdy-planet/mat1"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtSysBelt, needTemplate(tmpls, "topic/bdy-belt"), idxMapNames.Convert)
}

func emitBodies(wr io.Writer, matLs []string) (n int) {
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
				emitStar(wr, btStar, bdy)
			case gxy.Planet:
				if bdy.IsBelt() {
					btBelt.BindP(gxtSysBelt.Name, gxw.HtmlEsc(bdy.Name))
					btBelt.Bind(gxtSysBelt.Dist, gxm.Msg(wuiL7d, "%.2f", bdy.Dist))
					n += btBelt.Emit(wr)
				} else {
					emitPlanet(wr, btPlnt, btMat0, btMat1, matLs, bdy)
				}
			default:
				glog.Logf(l.Warn,
					"body without category in system %s",
					cmdr.Loc.System().Name())
			}
		}
	}
	return n
}

func emitStar(wr io.Writer, bt *gx.BounT, s *gxy.SysBody) (n int) {
	bt.BindP(gxtSysStar.Name, gxw.HtmlEsc(s.Name))
	bt.Bind(gxtSysStar.Dist, gxm.Msg(wuiL7d, "%.2f", s.Dist))
	bt.BindFmt(gxtSysStar.R, "%.2f", s.Radius)
	bt.Bind(gxtSysStar.RotPrd, webGuiTBD)
	bt.Bind(gxtSysStar.Tilt, webGuiTBD)
	n += bt.Emit(wr)
	return n
}

func emitPlanet(wr io.Writer, bt, btM0, btM1 *gx.BounT, matLs []string, p *gxy.SysBody) (n int) {
	bt.BindP(gxtSysPlanet.Name, gxw.HtmlEsc(p.Name))
	bt.Bind(gxtSysPlanet.Dist, gxm.Msg(wuiL7d, "%.2f", p.Dist))
	bt.BindP(gxtSysPlanet.Land, p.Landable)
	bt.Bind(gxtSysPlanet.R, gxm.Msg(wuiL7d, "%.2f", p.Radius))
	if len(p.Mats) == 0 {
		bt.Bind(gxtSysPlanet.Materials, gx.Empty)
	} else {
		btMats := gxtMatSec.NewBounT()
		bt.Bind(gxtSysPlanet.Materials, btMats)
		btMats.BindGen(gxtMatSec.Mats, func(wr io.Writer) (n int) {
			if len(p.Mats) == 0 {
				return 0
			} else {
				first := true
				for _, mat := range matLs {
					if f, ok := p.Mats[mat]; ok {
						nm, _ := nmMats.Map(mat)
						if first {
							btM0.BindP(gxtSysPMat0.Name, gxw.HtmlEsc(nm))
							btM0.Bind(gxtSysPMat0.Frac, gxm.Msg(wuiL7d, "%.2f", f))
							n += btM0.Emit(wr)
							first = false
						} else {
							btM1.BindP(gxtSysPMat1.Name, gxw.HtmlEsc(nm))
							btM1.Bind(gxtSysPMat1.Frac, gxm.Msg(wuiL7d, "%.2f", f))
							n += btM1.Emit(wr)
						}
					}
				}
			}
			return n
		})
	}
	n += bt.Emit(wr)
	return n
}

func wuiSys(w http.ResponseWriter, r *http.Request) {
	var matLs []string
	for mat, _ := range theGalaxy.Materials {
		matLs = append(matLs, mat)
	}
	sort.Slice(matLs,
		func(i, j int) bool { return cmprMatByL7d(matLs, i, j) })
	btEmit, btBind, hook := preparePage(dynSysStyles, gx.Empty, activeTopic(r))
	btFrame := gxtSysFrame.NewBounT()
	btBind.Bind(hook, btFrame)
	btFrame.BindGen(gxtSysFrame.Bodies, func(wr io.Writer) int {
		return emitBodies(wr, matLs)
	})
	btEmit.Emit(w)
}
