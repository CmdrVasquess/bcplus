package main

import (
	"io"
	"net/http"
	"sort"

	gxy "github.com/CmdrVasquess/BCplus/galaxy"
	gx "github.com/fractalqb/goxic"
	gxw "github.com/fractalqb/goxic/web"
	l "github.com/fractalqb/qblog"
)

var _ = &l.Logger{}

var gxtRescFrame struct {
	*gx.Template
	RawType []int `goxic:"rawmat-type"`
	RawMax  []int `goxic:"raw-max"`
	RawSum  []int `goxic:"raw-sum"`
	RawMats []int `goxic:"rawmats"`
	Man     []int `goxic:"list-man"`
	Enc     []int `goxic:"list-enc"`
}

var gxtRescList struct {
	*gx.Template
	Type  []int `goxic:"resc-type"`
	Sum   []int `goxic:"sum"`
	Max   []int `goxic:"max"`
	Items []int `goxic:"items"`
}

var gxtRescItem struct {
	*gx.Template
	Demand []int `goxic:"demand"`
	Grade  []int `goxic:"mat-grd"`
	XRef   []int `goxic:"xref"`
	Name   []int `goxic:"name"`
	Have   []int `goxic:"have"`
	Need   []int `goxic:"need"`
}

var gxtRawItem struct {
	*gx.Template
	Demand []int `goxic:"demand"`
	Grade  []int `goxic:"mat-grd"`
	XRef   []int `goxic:"xref"`
	Name   []int `goxic:"name"`
	Have   []int `goxic:"have"`
	Need   []int `goxic:"need"`
	Prc    []int `goxic:"raw-prc"`
	Body   []int `goxic:"raw-bdy"`
}

var dynRescStyles gx.Content

func loadRescTemplates() {
	tmpls := make(map[string]*gx.Template)
	if err := gxw.ParseHtmlTemplate(assetPath("resources.html"), "resources", tmpls); err != nil {
		panic("failed loading templates: " + err.Error())
	}
	dynRescStyles = pageLocalStyle(tmpls)
	gx.MustIndexMap(&gxtRescFrame, needTemplate(tmpls, "topic"))
	gx.MustIndexMap(&gxtRawItem, needTemplate(tmpls, "topic/rawmat"))
	gx.MustIndexMap(&gxtRescList, needTemplate(tmpls, "resc-list"))
	gx.MustIndexMap(&gxtRescItem, needTemplate(tmpls, "resc-list/item"))
}

func resourceCount(rescs CmdrsMats) int {
	res := 0
	for _, m := range rescs {
		res += int(m.Have)
	}
	return res
}

type bestRawMat struct {
	percent float32
	body    *gxy.SysBody
}

var rawSorted []string
var manSorted []string
var encSorted []string

func cmprMatByL7d(jnms []string, i, j int) bool {
	si := jnms[i]
	si, _ = nmMats.Map(si)
	sj := jnms[j]
	sj, _ = nmMats.Map(sj)
	return si < sj
}

func sortMats() {
	if len(rawSorted) > 0 && len(manSorted) > 0 && len(encSorted) > 0 {
		return
	}
	rawSorted, manSorted, encSorted = nil, nil, nil
	for _, mat := range theGalaxy.Materials {
		switch mat.Category {
		case 0:
			rawSorted = append(rawSorted, mat.JName)
		case 1:
			manSorted = append(manSorted, mat.JName)
		case 2:
			encSorted = append(encSorted, mat.JName)
		default:
			glog.Fatalf("material '%s' with unknown kind %d", mat.JName, mat.Category)
		}
	}
	sort.Slice(rawSorted,
		func(i, j int) bool { return cmprMatByL7d(rawSorted, i, j) })
	sort.Slice(manSorted,
		func(i, j int) bool { return cmprMatByL7d(manSorted, i, j) })
	sort.Slice(encSorted,
		func(i, j int) bool { return cmprMatByL7d(encSorted, i, j) })
}

func bestRawMats(ssys *gxy.StarSys) map[string]bestRawMat {
	res := make(map[string]bestRawMat)
	for _, bdy := range ssys.Bodies {
		if bdy.Mats != nil {
			for mat, prc := range bdy.Mats {
				if best, ok := res[mat]; ok {
					if prc > best.percent {
						res[mat] = bestRawMat{prc, bdy}
					} else if prc == best.percent && bdy.Dist > 0 && bdy.Dist < best.body.Dist {
						res[mat] = bestRawMat{prc, bdy}
					}
				} else {
					res[mat] = bestRawMat{prc, bdy}
				}
			}
		}
	}
	return res
}

func emitRawMats(mats CmdrsMats, wr io.Writer) (n int) {
	var best map[string]bestRawMat
	if theGame.Cmdr.Loc.Location != nil {
		best = bestRawMats(theGame.Cmdr.Loc.System())
	}
	btItem := gxtRawItem.NewBounT()
	for _, mat := range rawSorted {
		if m, ok := theGalaxy.Materials[mat]; !ok || m.Commons < 0 {
			btItem.BindP(gxtRawItem.Grade, "_")
		} else {
			btItem.BindP(gxtRawItem.Grade, m.Commons)
		}
		cmdrmat, cmdrHas := mats[mat]
		btItem.Bind(gxtRawItem.XRef, nmap(&nmMatsXRef, mat))
		btItem.Bind(gxtRawItem.Name, nmap(&nmMats, mat))
		if cmdrHas {
			if cmdrmat.Have == 0 {
				btItem.Bind(gxtRawItem.Have, gx.Empty)
			} else {
				btItem.BindP(gxtRawItem.Have, cmdrmat.Have)
			}
			if cmdrmat.Need == 0 {
				btItem.Bind(gxtRawItem.Demand, gx.Empty)
				btItem.Bind(gxtRawItem.Need, gx.Empty)
			} else {
				if cmdrmat.Have >= cmdrmat.Need {
					btItem.BindP(gxtRawItem.Demand, "engh")
				} else {
					btItem.BindP(gxtRawItem.Demand, "miss")
				}
				btItem.BindP(gxtRawItem.Need, cmdrmat.Need)
			}
		} else {
			btItem.Bind(gxtRawItem.Demand, gx.Empty)
			btItem.Bind(gxtRawItem.Have, gx.Empty)
			btItem.Bind(gxtRawItem.Need, gx.Empty)
		}
		if bm, ok := best[mat]; ok {
			btItem.BindFmt(gxtRawItem.Prc, "%.2f %%", bm.percent)
			btItem.Bind(gxtRawItem.Body, gxw.EscHtml{gx.Print{bm.body.Name}})
		} else {
			btItem.Bind(gxtRawItem.Prc, gx.Empty)
			btItem.Bind(gxtRawItem.Body, gx.Empty)
		}
		n += btItem.Emit(wr)
	}
	return n
}

func emitMatLs(mats []string, cMat CmdrsMats, name string, max int, wr io.Writer) (n int) {
	btLs := gxtRescList.NewBounT()
	btLs.Bind(gxtRescList.Type, gxw.EscHtml{nmap(&nmMatType, name)})
	btLs.BindP(gxtRescList.Sum, resourceCount(cMat))
	btLs.BindP(gxtRescList.Max, max)
	btLs.BindGen(gxtRescList.Items, func(wr io.Writer) (n int) {
		btItem := gxtRescItem.NewBounT()
		for _, mat := range mats {
			if m, ok := theGalaxy.Materials[mat]; !ok || m.Commons < 0 {
				btItem.BindP(gxtRescItem.Grade, "_")
			} else {
				btItem.BindFmt(gxtRescItem.Grade, "%d", m.Commons)
			}
			cmdrmat, cmdrHas := cMat[mat]
			btItem.Bind(gxtRescItem.XRef, nmap(&nmMatsXRef, mat))
			btItem.Bind(gxtRescItem.Name, nmap(&nmMats, mat))
			if cmdrHas {
				if cmdrmat.Have == 0 {
					btItem.Bind(gxtRescItem.Have, gx.Empty)
				} else {
					btItem.BindP(gxtRescItem.Have, cmdrmat.Have)
				}
				if cmdrmat.Need == 0 {
					btItem.Bind(gxtRescItem.Demand, gx.Empty)
					btItem.Bind(gxtRescItem.Need, gx.Empty)
				} else {
					if cmdrmat.Have >= cmdrmat.Need {
						btItem.BindP(gxtRescItem.Demand, "engh")
					} else {
						btItem.BindP(gxtRescItem.Demand, "miss")
					}
					btItem.BindP(gxtRescItem.Need, cmdrmat.Need)
				}
			} else {
				btItem.BindP(gxtRawItem.Demand, gx.Empty)
				btItem.Bind(gxtRescItem.Need, gx.Empty)
				btItem.Bind(gxtRescItem.Have, gx.Empty)
			}
			n += btItem.Emit(wr)
		}
		return n
	})
	n += btLs.Emit(wr)
	return n
}

func wuiResources(w http.ResponseWriter, r *http.Request) {
	btEmit, btBind, hook := preparePage(dynRescStyles)
	btFrame := gxtRescFrame.NewBounT()
	btBind.Bind(hook, btFrame)
	cmdr := &theGame.Cmdr
	haveRaw := resourceCount(cmdr.MatsRaw)
	haveMan := resourceCount(cmdr.MatsMan)
	btFrame.Bind(gxtRescFrame.RawType, gxw.EscHtml{nmap(&nmMatType, "raw")})
	btFrame.BindP(gxtRescFrame.RawMax, 1000-haveMan)
	btFrame.BindP(gxtRescFrame.RawSum, resourceCount(cmdr.MatsRaw))
	sortMats()
	btFrame.BindGen(gxtRescFrame.RawMats, func(wr io.Writer) int {
		return emitRawMats(cmdr.MatsRaw, wr)
	})
	btFrame.BindGen(gxtRescFrame.Man, func(wr io.Writer) int {
		return emitMatLs(manSorted, cmdr.MatsMan, "man", 1000-haveRaw, wr)
	})
	btFrame.BindGen(gxtRescFrame.Enc, func(wr io.Writer) int {
		return emitMatLs(encSorted, cmdr.MatsEnc, "enc", 500, wr)
	})
	btEmit.Emit(w)
}
