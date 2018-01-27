package main

import (
	"fmt"
	"io"
	"net/http"
	"sort"

	gxy "github.com/CmdrVasquess/BCplus/galaxy"
	gx "github.com/fractalqb/goxic"
	gxw "github.com/fractalqb/goxic/web"
	l "github.com/fractalqb/qblog"
)

var gxtSynFrame struct {
	*gx.Template
	Recipes []int
}

var gxtRecipe struct {
	*gx.Template
	RcpId     []int
	Name      []int
	Imprv     []int
	Builds    []int
	Materials []int
	Levels    []int
}

var gxtRcpBuild0 struct {
	*gx.Template
	Count []int
}

var gxtRcpBuildN struct {
	*gx.Template
	Level []int
	Count []int
}

var gxtHdrMat struct {
	*gx.Template
	Name     []int
	MatGrade []int
	Have     []int
}

var gxtMatLvl struct {
	*gx.Template
	Level     []int
	Bonus     []int
	Have      []int
	Need      []int
	Materials []int
}

var gxtLvlMatNone struct {
	*gx.Template
}

var gxtLvlMatGood struct {
	*gx.Template
	Count []int
}

var gxtLvlMatNeed struct {
	*gx.Template
	Count []int
}

var dynSynStyles gx.Content

func loadSynTemplates() {
	tmpls := make(map[string]*gx.Template)
	if err := gxw.ParseHtmlTemplate(assetPath("synth.html"), "synth", tmpls); err != nil {
		panic("failed loading templates: " + err.Error())
	}
	dynSynStyles = pgLocStyleFix(tmpls)
	//	endSynScrpit = pgEndScriptFix(tmpls)
	gx.MustIndexMap(&gxtSynFrame, needTemplate(tmpls, "topic"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtRecipe, needTemplate(tmpls, "topic/recipe"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtRcpBuild0, needTemplate(tmpls, "topic/recipe/build0"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtRcpBuildN, needTemplate(tmpls, "topic/recipe/buildN"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtHdrMat, needTemplate(tmpls, "topic/recipe/material"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtMatLvl, needTemplate(tmpls, "topic/recipe/level"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtLvlMatNone, needTemplate(tmpls, "topic/recipe/level/mat-none"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtLvlMatGood, needTemplate(tmpls, "topic/recipe/level/mat-good"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtLvlMatNeed, needTemplate(tmpls, "topic/recipe/level/mat-need"), idxMapNames.Convert)
}

func emitLevels(matLs []string, rcp *gxy.Synthesis, builds []int, wr io.Writer) (n int) {
	cmdr := &theGame.Cmdr
	btLvl := gxtMatLvl.NewBounT()
	btMatNo := gxtLvlMatNone.NewBounT()
	btMatGo := gxtLvlMatGood.NewBounT()
	btMatNd := gxtLvlMatNeed.NewBounT()
	for i, lvl := range rcp.Levels {
		btLvl.BindP(gxtMatLvl.Level, i+1)
		btLvl.BindP(gxtMatLvl.Bonus, lvl.Bonus)
		if builds[i] > 0 {
			btLvl.BindP(gxtMatLvl.Have, builds[i])
		} else {
			btLvl.Bind(gxtMatLvl.Have, gx.Empty)
		}
		btLvl.Bind(gxtMatLvl.Need, webGuiTBD)
		btLvl.BindGen(gxtMatLvl.Materials, func(wr io.Writer) (n int) {
			for _, mat := range matLs {
				var btMat *gx.BounT
				count, ok := lvl.Demand[mat]
				if ok && count > 0 {
					cmat := cmdr.Material(mat)
					if cmat == nil || cmat.Have <= 0 {
						btMatNd.BindP(gxtLvlMatNeed.Count, count)
						btMat = btMatNd
					} else if uint(cmat.Have) >= count {
						btMatGo.BindP(gxtLvlMatGood.Count, count)
						btMat = btMatGo
					} else {
						btMatNd.BindP(gxtLvlMatNeed.Count, count)
						btMat = btMatNd
					}
				} else {
					btMat = btMatNo
				}
				n += btMat.Emit(wr)
			}
			return n
		})
		n += btLvl.Emit(wr)
	}
	return n
}

func recipeBuilds(recipe *gxy.Synthesis) (res []int) {
	cmdr := &theGame.Cmdr
	for _, lvl := range recipe.Levels {
		build := -1
		for nmat, need := range lvl.Demand {
			cmat := cmdr.Material(nmat)
			if cmat == nil || cmat.Have <= int16(need) {
				build = 0
				break
			} else if tmp := uint(cmat.Have) / need; build < 0 || int(tmp) < build {
				build = int(tmp)
			}
		}
		if build < 0 {
			res = append(res, 0)
		} else {
			res = append(res, build)
		}
	}
	return res
}

func wuiSyn(w http.ResponseWriter, r *http.Request) {
	cmdr := &theGame.Cmdr
	btEmit, btBind, hook := preparePage(dynSynStyles, gx.Empty)
	btFrame := gxtSynFrame.NewBounT()
	btBind.Bind(hook, btFrame)
	btFrame.BindGen(gxtSynFrame.Recipes, func(wr io.Writer) (n int) {
		btRcp := gxtRecipe.NewBounT()
		btHdrMat := gxtHdrMat.NewBounT()
		for rcpid, recipe := range theGalaxy.Synth {
			var matSet = make(map[string]bool)
			for _, lvl := range recipe.Levels {
				for mat, _ := range lvl.Demand {
					matSet[mat] = true
				}
			}
			var matLs []string
			for mat, _ := range matSet {
				matLs = append(matLs, mat)
			}
			matSet = nil
			sort.Slice(matLs,
				func(i, j int) bool { return cmprMatByL7d(matLs, i, j) })
			btRcp.BindP(gxtRecipe.RcpId, rcpid)
			btRcp.BindP(gxtRecipe.Name, recipe.Name)
			btRcp.BindP(gxtRecipe.Imprv, recipe.Improves)
			builds := recipeBuilds(&recipe)
			btRcp.BindGen(gxtRecipe.Builds, func(wr io.Writer) (n int) {
				btBld0 := gxtRcpBuild0.NewBounT()
				btBldN := gxtRcpBuildN.NewBounT()
				for i, b := range builds {
					if i == 0 {
						btBld0.BindP(gxtRcpBuild0.Count, b)
						n += btBld0.Emit(wr)
					} else {
						btBldN.BindP(gxtRcpBuildN.Level, i+1)
						btBldN.BindP(gxtRcpBuildN.Count, b)
						n += btBldN.Emit(wr)
					}
				}
				return n
			})
			btRcp.BindGen(gxtRecipe.Materials, func(wr io.Writer) (n int) {
				for _, mat := range matLs {
					cmns := "_"
					if gxmat, ok := theGalaxy.Materials[mat]; ok {
						cmns = fmt.Sprintf("%d", gxmat.Commons)
					}
					name, _ := nmMats.Map(mat)
					btHdrMat.BindP(gxtHdrMat.Name, name)
					if cmat := cmdr.Material(mat); cmat == nil || cmat.Have == 0 {
						btHdrMat.BindP(gxtHdrMat.Have, 0)
						btHdrMat.BindP(gxtHdrMat.MatGrade, cmns)
					} else {
						btHdrMat.BindP(gxtHdrMat.Have, cmat.Have)
						btHdrMat.BindP(gxtHdrMat.MatGrade, cmns)
					}
					n += btHdrMat.Emit(wr)
				}
				return n
			})
			btRcp.BindGen(gxtRecipe.Levels, func(wr io.Writer) (n int) {
				return emitLevels(matLs, &recipe, builds, wr)
			})
			n += btRcp.Emit(wr)
		}
		return n
	})
	btEmit.Emit(w)
}

var synUsrOps = map[string]userHanlder{
	"mdmnd": synUsrOpMdmnd,
}

func synUsrOpMdmnd(gstat *GmState, evt map[string]interface{}) (reload bool) {
	rcpid, _ := attInt(evt, "rcpid")
	lvl, _ := attInt(evt, "level")
	count, _ := attInt(evt, "count")
	eulog.Logf(l.Debug, "synthesis set manual demand: id%d/%d=%d", rcpid, lvl, count)
	synth := &theGalaxy.Synth[rcpid]
	cmdr := &gstat.Cmdr
	cmdr.NeedSynth(synth, uint(lvl)-1, uint(count))
	return false
}
