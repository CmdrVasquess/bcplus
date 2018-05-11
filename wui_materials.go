package main

import (
	"io"
	"net/http"
	"sort"
	"strconv"

	"github.com/CmdrVasquess/BCplus/cmdr"
	c "github.com/CmdrVasquess/BCplus/cmdr"
	gxy "github.com/CmdrVasquess/BCplus/galaxy"
	gx "github.com/fractalqb/goxic"
	gxw "github.com/fractalqb/goxic/web"
	l "github.com/fractalqb/qblog"
)

var gxtRescFrame struct {
	*gx.Template
	ThSynrcps []int
	ThSynlvls []int
	Sections  []int
	FltHave   []int
	FltNeed   []int
}

var gxtThSynRcp struct {
	*gx.Template
	Name   []int
	Repeat []int
}

var gxtThSynLvl struct {
	*gx.Template
	Count []int
	Level []int
}

var gxtSecTitle struct {
	*gx.Template
	Cat      []int
	Category []int
	Have     []int
	Needs    []int
}

var gxtSecNeed struct {
	*gx.Template
	Need []int
}

var gxtSecRow struct {
	*gx.Template
	Cat      []int
	MatId    []int
	MatGrade []int
	Xref     []int
	Name     []int
	Have     []int
	Max      []int
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
	tpars := gxw.NewHtmlParser()
	if err := tpars.ParseFile(assetPath("materials.html"), "resources", tmpls); err != nil {
		panic("failed loading templates: " + err.Error())
	}
	dynRescStyles = pgLocStyleFix(tmpls)
	endRescScript = pgEndScript(tmpls)
	gx.MustIndexMap(&gxtRescFrame, needTemplate(tmpls, "topic"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtThSynRcp, needTemplate(tmpls, "topic/th-synrcp"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtThSynLvl, needTemplate(tmpls, "topic/th-synlvl"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtSecTitle, needTemplate(tmpls, "topic/sec-title"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtSecNeed, needTemplate(tmpls, "topic/sec-title/need"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtSecRow, needTemplate(tmpls, "topic/sec-row"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtRowSrc1, needTemplate(tmpls, "topic/sec-row/src1"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtRowSrc2, needTemplate(tmpls, "topic/sec-row/src2"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtRowNeed, needTemplate(tmpls, "topic/sec-row/need"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtHideCat, needTemplate(tmpls, "end-script/hide-cat"), idxMapNames.Convert)
}

func resourceCount(rescs cmdr.CmdrsMats) (have, need int) {
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

func emitRawMats(wr io.Writer, bt *gx.BounT, cmdr *cmdr.Commander, ndSyn []cmdr.SynthRef) (n int) {
	mats := cmdr.MatsRaw
	var best map[string]bestRawMat
	if theGame.Cmdr.Loc.Ref != nil {
		best = bestRawMats(theGame.Cmdr.Loc.System())
	}
	btSrc := gxtRowSrc2.NewBounT(nil)
	bt.Bind(gxtSecRow.Source, btSrc)
	bt.BindP(gxtSecRow.ManIdx, 1)
	btNeed := gxtRowNeed.NewBounT(nil)
	for _, mat := range rawSorted {
		bt.Bind(gxtSecRow.MatId, nmap(&nmMatsId, mat))
		if m, ok := theGalaxy.Materials[mat]; !ok || m.Commons < 0 {
			bt.BindP(gxtSecRow.MatGrade, "_")
			bt.Bind(gxtSecRow.Max, webGuiTBD)
		} else {
			bt.BindP(gxtSecRow.MatGrade, m.Commons)
			max, _ := nmMatGrade.Map(strconv.Itoa(int(m.Commons)), nmMGrdRaw)
			bt.BindP(gxtSecRow.Max, max)
		}
		bt.Bind(gxtSecRow.Xref, nmap(&nmMatsXRef, mat))
		bt.Bind(gxtSecRow.Name, nmap(&nmMats, mat))
		cmdrmat, cmdrHas := mats[mat]
		if cmdrHas {
			if cmdrmat.Have == 0 {
				bt.Bind(gxtSecRow.Have, webGuiNOC)
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
			bt.Bind(gxtSecRow.Have, webGuiNOC)
			bt.Bind(gxtSecRow.Need, gx.Empty)
		}
		if bm, ok := best[mat]; ok {
			btSrc.BindFmt(gxtRowSrc2.Val1, "%.2f %%", bm.percent)
			btSrc.Bind(gxtRowSrc2.Val2, gxw.HtmlEsc{gx.Print{bm.body.Name}})
		} else {
			btSrc.Bind(gxtRowSrc2.Val1, gx.Empty)
			btSrc.Bind(gxtRowSrc2.Val2, gx.Empty)
		}
		bt.BindGen(gxtSecRow.Needs, func(wr io.Writer) (n int) {
			for _, sr := range ndSyn {
				if rcpNo := cmdr.Synth[sr]; rcpNo > 0 {
					rcp, lvl := sr.Get()
					rl := rcp.Levels[lvl]
					if matNo, _ := rl.Demand[mat]; matNo > 0 {
						btNeed.BindP(gxtRowNeed.Count, rcpNo*matNo)
					} else {
						btNeed.Bind(gxtRowNeed.Count, gx.Empty)
					}
				} else {
					btNeed.Bind(gxtRowNeed.Count, gx.Empty)
				}
				n += btNeed.Emit(wr)
			}
			return n
		})
		n += bt.Emit(wr)
	}
	return n
}

func emitMatLs(
	wr io.Writer,
	bt, src *gx.BounT,
	cat string,
	mats []string,
	cmdr *cmdr.Commander,
	cMat cmdr.CmdrsMats,
	ndSyn []cmdr.SynthRef) (n int) {
	src.Bind(gxtRowSrc1.Value, webGuiTBD)
	bt.BindP(gxtSecRow.ManIdx, 0)
	btNeed := gxtRowNeed.NewBounT(nil)
	for _, mat := range mats {
		bt.Bind(gxtSecRow.MatId, nmap(&nmMatsId, mat))
		if m, ok := theGalaxy.Materials[mat]; !ok || m.Commons < 0 {
			bt.BindP(gxtSecRow.MatGrade, "_")
			bt.Bind(gxtSecRow.Max, webGuiTBD)
		} else {
			bt.BindP(gxtSecRow.MatGrade, m.Commons)
			max, _ := nmMatGrade.MapNm(strconv.Itoa(int(m.Commons)), cat)
			bt.BindP(gxtSecRow.Max, max)
		}
		bt.Bind(gxtSecRow.Xref, nmap(&nmMatsXRef, mat))
		bt.Bind(gxtSecRow.Name, nmap(&nmMats, mat))
		cmdrmat, cmdrHas := cMat[mat]
		if cmdrHas {
			if cmdrmat.Have == 0 {
				bt.Bind(gxtSecRow.Have, webGuiNOC)
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
			bt.Bind(gxtSecRow.Have, webGuiNOC)
		}
		bt.BindGen(gxtSecRow.Needs, func(wr io.Writer) (n int) {
			for _, sr := range ndSyn {
				if rcpNo := cmdr.Synth[sr]; rcpNo > 0 {
					rcp, lvl := sr.Get()
					rl := rcp.Levels[lvl]
					if matNo, _ := rl.Demand[mat]; matNo > 0 {
						btNeed.BindP(gxtRowNeed.Count, rcpNo*matNo)
					} else {
						btNeed.Bind(gxtRowNeed.Count, gx.Empty)
					}
				} else {
					btNeed.Bind(gxtRowNeed.Count, gx.Empty)
				}
				n += btNeed.Emit(wr)
			}
			return n
		})
		n += bt.Emit(wr)
	}
	return n
}

func cmdrNeedsSynths(cmdr *cmdr.Commander) (res []cmdr.SynthRef) {
	for rIdx, _ := range theGalaxy.Synth {
		recipe := &theGalaxy.Synth[rIdx]
		for lvl := 0; lvl < len(recipe.Levels); lvl++ {
			ndKey := c.MkSynthRef(recipe, lvl)
			if ndNo, _ := cmdr.Synth[ndKey]; ndNo > 0 {
				res = append(res, ndKey)
			}
		}
	}
	return res
}

func needsHdrs(wr io.Writer, cmdr *cmdr.Commander, needSynths []cmdr.SynthRef) (n int) {
	btThSyn := gxtThSynRcp.NewBounT(nil)
	var j int
	for i := 0; i < len(needSynths); i = j {
		snm, _ := needSynths[i].Split()
		j = i + 1
		for j < len(needSynths) {
			enm, _ := needSynths[j].Split()
			if enm != snm {
				break
			}
			j++
		}
		btThSyn.BindP(gxtThSynRcp.Name, gxw.EscHtml(snm))
		btThSyn.BindP(gxtThSynRcp.Repeat, j-i)
		n += btThSyn.Emit(wr)
	}
	return n
}

func needsLvls(wr io.Writer, cmdr *cmdr.Commander, needSynths []cmdr.SynthRef) (n int) {
	btThLvl := gxtThSynLvl.NewBounT(nil)
	for _, sr := range needSynths {
		count := cmdr.Synth[sr]
		_, lvl := sr.Split()
		btThLvl.BindP(gxtThSynLvl.Count, count)
		btThLvl.Bind(gxtThSynLvl.Level, nmap(&nmSynthLvl, strconv.Itoa(lvl)))
		n += btThLvl.Emit(wr)
	}
	return n
}

func secTitle(bt *gx.BounT, wr io.Writer, cat string, have, need, needs int) (n int) {
	catNm, _ := nmMatType.Map(cat)
	bt.BindP(gxtSecTitle.Cat, cat)
	bt.Bind(gxtSecTitle.Category, gxw.HtmlEsc{gx.Print{catNm}})
	bt.BindP(gxtSecTitle.Have, have)
	btSecNeed := gxtSecNeed.NewInitBounT(gx.Empty, nil)
	bt.BindGen(gxtSecTitle.Needs, func(wr io.Writer) (n int) {
		for needs > 0 {
			n += btSecNeed.Emit(wr)
			needs--
		}
		return n
	})
	n += bt.Emit(wr)
	return n
}

func wuiMats(w http.ResponseWriter, r *http.Request) {
	btEndScript := endRescScript.NewBounT(nil)
	btEndScript.BindGenName("hide-cats", func(wr io.Writer) (n int) {
		btHide := gxtHideCat.NewBounT(nil)
		for cat, doHide := range theGame.MatCatHide {
			if doHide {
				btHide.BindP(gxtHideCat.Cat, cat)
				n += btHide.Emit(wr)
			}
		}
		return n
	})
	btEmit, btBind, hook := preparePage(dynRescStyles, gx.Empty, btEndScript, activeTopic(r))
	btFrame := gxtRescFrame.NewBounT(nil)
	btBind.Bind(hook, btFrame)
	cmdr := &theGame.Cmdr
	haveRaw, needRaw := resourceCount(cmdr.MatsRaw)
	haveMan, needMan := resourceCount(cmdr.MatsMan)
	haveEnc, needEnc := resourceCount(cmdr.MatsEnc)
	sortMats()
	needSynths := cmdrNeedsSynths(cmdr)
	btFrame.BindGen(gxtRescFrame.ThSynrcps, func(wr io.Writer) (n int) {
		n = needsHdrs(wr, cmdr, needSynths)
		return n
	})
	btFrame.BindGen(gxtRescFrame.ThSynlvls, func(wr io.Writer) (n int) {
		n = needsLvls(wr, cmdr, needSynths)
		return n
	})
	if len(theGame.MatFlt.Have) > 0 {
		btFrame.BindP(gxtRescFrame.FltHave, theGame.MatFlt.Have)
		btFrame.BindP(gxtRescFrame.FltNeed, theGame.MatFlt.Need)
	} else {
		btFrame.BindP(gxtRescFrame.FltHave, "alor")
		btFrame.BindP(gxtRescFrame.FltNeed, true)
	}
	btSec := gxtSecTitle.NewBounT(nil)
	btRow := gxtSecRow.NewBounT(nil)
	btSrc1 := gxtRowSrc1.NewBounT(nil)
	btFrame.BindGen(gxtRescFrame.Sections, func(wr io.Writer) (n int) {
		n += secTitle(btSec, wr, "raw", haveRaw, needRaw, len(needSynths))
		n += emitRawMats(wr, btRow, cmdr, needSynths)
		btRow.Bind(gxtSecRow.Source, btSrc1)
		n += secTitle(btSec, wr, "man", haveMan, needMan, len(needSynths))
		n += emitMatLs(wr, btRow, btSrc1, "man", manSorted, cmdr, cmdr.MatsMan, needSynths)
		n += secTitle(btSec, wr, "enc", haveEnc, needEnc, len(needSynths))
		n += emitMatLs(wr, btRow, btSrc1, "enc", encSorted, cmdr, cmdr.MatsEnc, needSynths)
		return n
	})
	btEmit.Emit(w)
}

var matUsrOps = map[string]userHanlder{
	"vis":   matUsrOpCatVis,
	"mdmnd": matUsrOpMdmnd,
	"mflt":  matUsrFilter,
}

func matUsrOpCatVis(gstat *cmdr.GmState, evt map[string]interface{}) (reload bool) {
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

func matUsrOpMdmnd(gstat *c.GmState, evt map[string]interface{}) (reload bool) {
	mat, _ := attStr(evt, "matid")
	mat, _ = nmMatsIdRev.Map(mat)
	count, _ := attInt(evt, "count")
	eulog.Logf(l.Debug, "materials set manual demand: %s=%d", mat, count)
	cmdr := &gstat.Cmdr
	var matMap c.CmdrsMats
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
		cmat = &c.Material{
			Have: 0,
			Need: int16(count),
		}
		matMap[mat] = cmat
	} else {
		cmat.Need = int16(count)
	}
	return false
}

func matUsrFilter(gstat *c.GmState, evt map[string]interface{}) (reload bool) {
	gstat.MatFlt.Have, _ = attStr(evt, "have")
	gstat.MatFlt.Need, _ = attBool(evt, "need")
	return false
}
