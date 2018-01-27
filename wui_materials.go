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

type MatFilter struct {
	Have string
	Need bool
}

var gxtRescFrame struct {
	*gx.Template
	ThNeeds  []int
	Sections []int
	FltHave  []int
	FltNeed  []int
}

var gxtThNeed struct {
	*gx.Template
	Need []int
}

var gxtSecTitle struct {
	*gx.Template
	Cat      []int
	Category []int
	Have     []int
	Free     []int
	Needs    []int
}

var gxtSecRow struct {
	*gx.Template
	Cat      []int
	MatId    []int
	MatGrade []int
	Xref     []int
	Name     []int
	Have     []int
	Need     []int
	Source   []int
	Needs    []int
	ManIdx   []int
}

var gxtRowSrc1 struct {
	*gx.Template
	Value []int
}

var gxtRowSrc2 struct {
	*gx.Template
	Val1 []int
	Val2 []int
}

var gxtRowNeed struct {
	*gx.Template
	Count []int
}

var gxtHideCat struct {
	*gx.Template
	Cat []int
}

var dynRescStyles gx.Content
var endRescScript *gx.Template

func loadRescTemplates() {
	tmpls := make(map[string]*gx.Template)
	if err := gxw.ParseHtmlTemplate(assetPath("materials.html"), "resources", tmpls); err != nil {
		panic("failed loading templates: " + err.Error())
	}
	dynRescStyles = pgLocStyleFix(tmpls)
	endRescScript = pgEndScript(tmpls)
	gx.MustIndexMap(&gxtRescFrame, needTemplate(tmpls, "topic"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtThNeed, needTemplate(tmpls, "topic/th-need"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtSecTitle, needTemplate(tmpls, "topic/sec-title"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtSecRow, needTemplate(tmpls, "topic/sec-row"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtRowSrc1, needTemplate(tmpls, "topic/sec-row/src1"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtRowSrc2, needTemplate(tmpls, "topic/sec-row/src2"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtRowNeed, needTemplate(tmpls, "topic/sec-row/need"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtHideCat, needTemplate(tmpls, "end-script/hide-cat"), idxMapNames.Convert)
}

func resourceCount(rescs CmdrsMats) (have, need int) {
	for _, m := range rescs {
		have += int(m.Have)
		need += int(m.Need)
	}
	return have, need
}

type bestRawMat struct {
	percent float32
	body    *gxy.SysBody
}

var rawSorted []string
var manSorted []string
var encSorted []string

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

func emitRawMats(wr io.Writer, bt *gx.BounT, mats CmdrsMats) (n int) {
	var best map[string]bestRawMat
	if theGame.Cmdr.Loc.Location != nil {
		best = bestRawMats(theGame.Cmdr.Loc.System())
	}
	btSrc := gxtRowSrc2.NewBounT()
	bt.Bind(gxtSecRow.Source, btSrc)
	bt.BindP(gxtSecRow.ManIdx, 1)
	for _, mat := range rawSorted {
		bt.Bind(gxtSecRow.MatId, nmap(&nmMatsId, mat))
		if m, ok := theGalaxy.Materials[mat]; !ok || m.Commons < 0 {
			bt.BindP(gxtSecRow.MatGrade, "_")
		} else {
			bt.BindP(gxtSecRow.MatGrade, m.Commons)
		}
		bt.Bind(gxtSecRow.Xref, nmap(&nmMatsXRef, mat))
		bt.Bind(gxtSecRow.Name, nmap(&nmMats, mat))
		cmdrmat, cmdrHas := mats[mat]
		if cmdrHas {
			if cmdrmat.Have == 0 {
				bt.Bind(gxtSecRow.Have, gx.Empty)
			} else {
				bt.BindP(gxtSecRow.Have, cmdrmat.Have)
			}
			bt.BindP(gxtSecRow.Cat, "raw")
			if cmdrmat.Need == 0 {
				bt.Bind(gxtSecRow.Need, gx.Empty)
			} else {
				bt.BindP(gxtSecRow.Need, cmdrmat.Need)
			}
		} else {
			bt.BindP(gxtSecRow.Cat, "raw")
			bt.Bind(gxtSecRow.Have, gx.Empty)
			bt.Bind(gxtSecRow.Need, gx.Empty)
		}
		if bm, ok := best[mat]; ok {
			btSrc.BindFmt(gxtRowSrc2.Val1, "%.2f %%", bm.percent)
			btSrc.Bind(gxtRowSrc2.Val2, gxw.EscHtml{gx.Print{bm.body.Name}})
		} else {
			btSrc.Bind(gxtRowSrc2.Val1, gx.Empty)
			btSrc.Bind(gxtRowSrc2.Val2, gx.Empty)
		}
		n += bt.Emit(wr)
	}
	return n
}

func emitMatLs(wr io.Writer, bt, src *gx.BounT, cat string, mats []string, cMat CmdrsMats) (n int) {
	src.Bind(gxtRowSrc1.Value, webGuiTBD)
	bt.BindP(gxtSecRow.ManIdx, 0)
	for _, mat := range mats {
		bt.Bind(gxtSecRow.MatId, nmap(&nmMatsId, mat))
		if m, ok := theGalaxy.Materials[mat]; !ok || m.Commons < 0 {
			bt.BindP(gxtSecRow.MatGrade, "_")
		} else {
			bt.BindP(gxtSecRow.MatGrade, m.Commons)
		}
		bt.Bind(gxtSecRow.Xref, nmap(&nmMatsXRef, mat))
		bt.Bind(gxtSecRow.Name, nmap(&nmMats, mat))
		cmdrmat, cmdrHas := cMat[mat]
		if cmdrHas {
			if cmdrmat.Have == 0 {
				bt.Bind(gxtSecRow.Have, gx.Empty)
			} else {
				bt.BindP(gxtSecRow.Have, cmdrmat.Have)
			}
			bt.BindP(gxtSecRow.Cat, cat)
			if cmdrmat.Need == 0 {
				bt.Bind(gxtSecRow.Need, gx.Empty)
			} else {
				bt.BindP(gxtSecRow.Need, cmdrmat.Need)
			}
		} else {
			bt.BindP(gxtSecRow.Cat, cat)
			bt.Bind(gxtSecRow.Need, gx.Empty)
			bt.Bind(gxtSecRow.Have, gx.Empty)
		}
		n += bt.Emit(wr)
	}
	return n
}

func secTitle(bt *gx.BounT, wr io.Writer, cat string, have, need, free, needs int) (n int) {
	catNm, _ := nmMatType.Map(cat)
	bt.BindP(gxtSecTitle.Cat, cat)
	bt.Bind(gxtSecTitle.Category, gxw.EscHtml{gx.Print{catNm}})
	bt.BindP(gxtSecTitle.Have, have)
	bt.BindP(gxtSecTitle.Free, free)
	bt.Bind(gxtSecTitle.Needs, gx.Empty)
	n += bt.Emit(wr)
	return n
}

func wuiMats(w http.ResponseWriter, r *http.Request) {
	btEndScript := endRescScript.NewBounT()
	btEndScript.BindGenName("hide-cats", func(wr io.Writer) (n int) {
		btHide := gxtHideCat.NewBounT()
		for cat, doHide := range theGame.MatCatHide {
			if doHide {
				btHide.BindP(gxtHideCat.Cat, cat)
				n += btHide.Emit(wr)
			}
		}
		return n
	})
	btEmit, btBind, hook := preparePage(dynRescStyles, btEndScript)
	btFrame := gxtRescFrame.NewBounT()
	btBind.Bind(hook, btFrame)
	cmdr := &theGame.Cmdr
	haveRaw, needRaw := resourceCount(cmdr.MatsRaw)
	haveMan, needMan := resourceCount(cmdr.MatsMan)
	haveEnc, needEnc := resourceCount(cmdr.MatsEnc)
	sortMats()
	//	btThNeed := gxtThNeed.NewBounT()
	//	btThNeed.Bind(gxtThNeed.Need, webGuiNOC)
	btFrame.Bind(gxtRescFrame.ThNeeds, gx.Empty)
	if len(theGame.MatFlt.Have) > 0 {
		btFrame.BindP(gxtRescFrame.FltHave, theGame.MatFlt.Have)
		btFrame.BindP(gxtRescFrame.FltNeed, theGame.MatFlt.Need)
	} else {
		btFrame.BindP(gxtRescFrame.FltHave, "alor")
		btFrame.BindP(gxtRescFrame.FltNeed, true)
	}
	btSec := gxtSecTitle.NewBounT()
	btRow := gxtSecRow.NewBounT()
	btRow.Bind(gxtSecRow.Needs, gx.Empty)
	btSrc1 := gxtRowSrc1.NewBounT()
	btFrame.BindGen(gxtRescFrame.Sections, func(wr io.Writer) (n int) {
		rawManFree := 1000 - haveRaw - haveMan
		n += secTitle(btSec, wr, "raw", haveRaw, needRaw, rawManFree, 1)
		n += emitRawMats(wr, btRow, cmdr.MatsRaw)
		btRow.Bind(gxtSecRow.Source, btSrc1)
		n += secTitle(btSec, wr, "man", haveMan, needMan, rawManFree, 1)
		n += emitMatLs(wr, btRow, btSrc1, "man", manSorted, cmdr.MatsMan)
		n += secTitle(btSec, wr, "enc", haveEnc, needEnc, 500-haveEnc, 1)
		n += emitMatLs(wr, btRow, btSrc1, "enc", encSorted, cmdr.MatsEnc)
		return n
	})
	btEmit.Emit(w)
}

var matUsrOps = map[string]userHanlder{
	"vis":   matUsrOpCatVis,
	"mdmnd": matUsrOpMdmnd,
	"mflt":  matUsrFilter,
}

func matUsrOpCatVis(gstat *GmState, evt map[string]interface{}) (reload bool) {
	cat, ok := attStr(evt, "cat")
	if !ok {
		eulog.Log(l.Error, "materials visibility changes has no category")
		return false
	}
	vis, ok := attStr(evt, "vis")
	if !ok {
		eulog.Log(l.Error, "materials visibility changes has no visibility")
		return false
	}
	gstat.MatCatHide[cat] = vis == "collapse"
	return false
}

func matUsrOpMdmnd(gstat *GmState, evt map[string]interface{}) (reload bool) {
	mat, _ := attStr(evt, "matid")
	mat, _ = nmMatsIdRev.Map(mat)
	count, _ := attInt(evt, "count")
	eulog.Logf(l.Debug, "materials set manual demand: %s=%d", mat, count)
	cmdr := &gstat.Cmdr
	var matMap CmdrsMats
	switch theGalaxy.Materials[mat].Category {
	case gxy.Raw:
		matMap = cmdr.MatsRaw
	case gxy.Man:
		matMap = cmdr.MatsMan
	case gxy.Enc:
		matMap = cmdr.MatsEnc
	}
	cmat, _ := matMap[mat]
	if cmat == nil {
		cmat = &Material{
			Have: 0,
			Need: int16(count),
		}
		matMap[mat] = cmat
	} else {
		cmat.Need = int16(count)
	}
	return false
}

func matUsrFilter(gstat *GmState, evt map[string]interface{}) (reload bool) {
	gstat.MatFlt.Have, _ = attStr(evt, "have")
	gstat.MatFlt.Need, _ = attBool(evt, "need")
	return false
}
