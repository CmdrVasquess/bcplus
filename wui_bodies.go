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

var gxtBdyFrame struct {
	*gx.Template
	Bodies []int
}

var gxtBdyTerence struct {
	*gx.Template
}

var gxtBdyStar struct {
	*gx.Template
	Type   []int
	Name   []int
	Dist   []int
	R      []int
	RotPrd []int
	Tilt   []int
	Rings  []int
}

var gxtBdyPlanet struct {
	*gx.Template
	Type      []int
	Name      []int
	Dist      []int
	Land      []int
	R         []int
	Materials []int
	Rings     []int
}

var gxtMatSec struct {
	*gx.Template
	Mats []int
}

var gxtBdyPMat0 struct {
	*gx.Template
	Class   []int
	Name    []int
	Frac    []int
	Matitle []int
}

var gxtBdyPMatN struct {
	*gx.Template
	Class   []int
	Name    []int
	Frac    []int
	Matitle []int
}

var gxtBdyBelt struct {
	*gx.Template
	Name []int
	Dist []int
}

var gxtRingSec struct {
	*gx.Template
}

var dynBdyStyles gx.Content

func loadBdyTemplates() {
	tmpls := make(map[string]*gx.Template)
	tpars := gxw.NewHtmlParser()
	if err := tpars.ParseFile(assetPath("bodies.html"), "bodies", tmpls); err != nil {
		panic("failed loading templates: " + err.Error())
	}
	dynBdyStyles = pgLocStyleFix(tmpls)
	//	endShpScrpit = pgEndScriptFix(tmpls)
	gx.MustIndexMap(&gxtBdyFrame, needTemplate(tmpls, "topic"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtBdyTerence, needTemplate(tmpls, "topic/bdy-none"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtBdyStar, needTemplate(tmpls, "topic/bdy-star"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtBdyPlanet, needTemplate(tmpls, "topic/bdy-planet"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtMatSec, needTemplate(tmpls, "topic/bdy-planet/mat-section"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtBdyPMat0, needTemplate(tmpls, "topic/bdy-planet/mat0"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtBdyPMatN, needTemplate(tmpls, "topic/bdy-planet/mat1"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtRingSec, needTemplate(tmpls, "topic/ring-section"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtBdyBelt, needTemplate(tmpls, "topic/bdy-belt"), idxMapNames.Convert)
}

func emitBodies(wr io.Writer, matLs []string) (n int) {
	cmdr := &theGame.Cmdr
	if cmdr.Loc.Nil() || len(cmdr.Loc.System().Bodies) == 0 {
		btNone := gxtBdyTerence.NewInitBounT(webGuiTBD, nil)
		n += btNone.Emit(wr)
	} else {
		bdys := make([]*gxy.SysBody, len(cmdr.Loc.System().Bodies))
		copy(bdys, cmdr.Loc.System().Bodies)
		sort.Slice(bdys, func(i, j int) bool {
			di, dj := bdys[i].Dist, bdys[j].Dist
			return di < dj
		})
		btStar := gxtBdyStar.NewBounT(nil)
		btPlnt := gxtBdyPlanet.NewBounT(nil)
		btBelt := gxtBdyBelt.NewBounT(nil)
		btMat0 := gxtBdyPMat0.NewBounT(nil)
		btMat1 := gxtBdyPMatN.NewBounT(nil)
		for _, bdy := range bdys {
			switch bdy.Cat {
			case gxy.Star:
				emitStar(wr, btStar, bdy)
			case gxy.Planet:
				if bdy.IsBelt() {
					btBelt.BindP(gxtBdyBelt.Name, gxw.EscHtml(bdy.Name))
					btBelt.Bind(gxtBdyBelt.Dist, gxm.Msg(wuiL7d, "%.2f", bdy.Dist))
					n += btBelt.Emit(wr)
				} else {
					emitPlanet(wr, btPlnt, btMat0, btMat1, matLs, bdy)
				}
			default:
				glog.Logf(l.Warn,
					"body without category in system %s: %s",
					cmdr.Loc.System().Name(),
					bdy.Name)
			}
		}
	}
	return n
}

func emitStar(wr io.Writer, bt *gx.BounT, s *gxy.SysBody) (n int) {
	bt.BindP(gxtBdyStar.Type, gxw.EscHtml(s.Type))
	bt.BindP(gxtBdyStar.Name, gxw.EscHtml(s.Name))
	bt.Bind(gxtBdyStar.Dist, gxm.Msg(wuiL7d, "%.2f", s.Dist))
	bt.BindFmt(gxtBdyStar.R, "%.2f", s.Radius)
	bt.Bind(gxtBdyStar.RotPrd, webGuiTBD)
	bt.Bind(gxtBdyStar.Tilt, webGuiTBD)
	// TODO Rings
	bt.Bind(gxtBdyStar.Rings, gx.Empty)
	n += bt.Emit(wr)
	return n
}

func emitPlanet(wr io.Writer, bt, btM0, btMN *gx.BounT, matLs []string, p *gxy.SysBody) (n int) {
	bt.BindP(gxtBdyPlanet.Type, gxw.EscHtml(p.Type))
	bt.BindP(gxtBdyPlanet.Name, gxw.EscHtml(p.Name))
	bt.Bind(gxtBdyPlanet.Dist, gxm.Msg(wuiL7d, "%.2f", p.Dist))
	bt.BindP(gxtBdyPlanet.Land, p.Landable)
	bt.Bind(gxtBdyPlanet.R, gxm.Msg(wuiL7d, "%.2f", p.Radius))
	if len(p.Mats) == 0 {
		bt.Bind(gxtBdyPlanet.Materials, gx.Empty)
	} else {
		btMats := gxtMatSec.NewBounT(nil)
		bt.Bind(gxtBdyPlanet.Materials, btMats)
		btMats.BindGen(gxtMatSec.Mats, func(wr io.Writer) (n int) {
			if len(p.Mats) == 0 {
				return 0
			} else {
				cmdr := &theGame.Cmdr
				first := true
				for _, mat := range matLs {
					if f, ok := p.Mats[mat]; ok {
						cls := ""
						if need := cmdr.NeedsMat(mat); need > 0 {
							cmat := cmdr.Material(mat)
							if need > int(cmat.Have) {
								cls = "miss"
							} else {
								cls = "engh"
							}
							if first {
								btM0.BindFmt(gxtBdyPMat0.Matitle, "%d / %d",
									cmat.Have,
									need)
							} else {
								btMN.BindFmt(gxtBdyPMatN.Matitle, "%d / %d",
									cmat.Have,
									need)
							}
						} else {
							if first {
								btM0.Bind(gxtBdyPMat0.Matitle, gx.Empty)
							} else {
								btMN.Bind(gxtBdyPMatN.Matitle, gx.Empty)
							}
						}
						nm, _ := nmMats.Map(mat)
						if first {
							btM0.BindP(gxtBdyPMat0.Class, cls)
							btM0.BindP(gxtBdyPMat0.Name, gxw.EscHtml(nm))
							btM0.Bind(gxtBdyPMat0.Frac, gxm.Msg(wuiL7d, "%.2f", f))
							n += btM0.Emit(wr)
							first = false
						} else {
							btMN.BindP(gxtBdyPMatN.Class, cls)
							btMN.BindP(gxtBdyPMatN.Name, gxw.EscHtml(nm))
							btMN.Bind(gxtBdyPMatN.Frac, gxm.Msg(wuiL7d, "%.2f", f))
							n += btMN.Emit(wr)
						}
					}
				}
			}
			return n
		})
	}
	// TODO rings
	bt.Bind(gxtBdyPlanet.Rings, gx.Empty)
	n += bt.Emit(wr)
	return n
}

func wuiBdys(w http.ResponseWriter, r *http.Request) {
	var matLs []string
	for mat, _ := range theGalaxy.Materials {
		matLs = append(matLs, mat)
	}
	sort.Slice(matLs,
		func(i, j int) bool { return cmprMatByL7d(matLs, i, j) })
	btEmit, btBind, hook := preparePage(dynBdyStyles, gx.Empty, gx.Empty, activeTopic(r))
	btFrame := gxtBdyFrame.NewBounT(nil)
	btBind.Bind(hook, btFrame)
	btFrame.BindGen(gxtBdyFrame.Bodies, func(wr io.Writer) int {
		return emitBodies(wr, matLs)
	})
	btEmit.Emit(w)
}
