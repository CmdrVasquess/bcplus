package main

import (
	"io"
	"math"
	"net/http"
	"strings"
	str "strings"
	"time"

	c "github.com/CmdrVasquess/BCplus/cmdr"
	gxy "github.com/CmdrVasquess/BCplus/galaxy"
	gx "github.com/fractalqb/goxic"
	gxm "github.com/fractalqb/goxic/textmessage"
	gxw "github.com/fractalqb/goxic/web"
	l "github.com/fractalqb/qblog"
)

var gxtTrvlFrame struct {
	*gx.Template
	AvgHeads []int
	AvgRows  []int
	Dests    []int `goxic:"destinations"`
	LypjMax  []int
	LypjAvg  []int
	ShipOpts []int `goxic:"shipopts"`
}

var gxtShipOpt struct {
	*gx.Template
	Id   []int
	Ship []int
	Jump []int
}

var gxtShipOptSel struct {
	*gx.Template
	Id   []int
	Ship []int
	Jump []int
}

var gxtTrvlAvgHead struct {
	*gx.Template
	Num []int `goxic:"count"`
}

var gxtTrvlAvgRow struct {
	*gx.Template
	Title []int
	Vals  []int `goxic:"values"`
}

var gxtTrvlAvgVal1 struct {
	*gx.Template
	Data []int
}

var gxtTrvlAvgVal2 struct {
	*gx.Template
	Sum []int
	Avg []int
}

var gxtTrvlDestRow struct {
	*gx.Template
	DstId    []int
	Name     []int
	HomeFlag []int
	Dist     []int
	ETD      []int `goxic:"etd"`
	EJD      []int `goxic:"ejd"`
	CooX     []int `goxic:"coox"`
	CooY     []int `goxic:"cooy"`
	CooZ     []int `goxic:"cooz"`
	Note     []int
	Tags     []int
}

var dynTrvlStyles gx.Content
var endTrvlScrpit gx.Content

func loadTrvlTemplates() {
	tmpls := make(map[string]*gx.Template)
	if err := gxw.ParseHtmlTemplate(assetPath("travel.html"), "travel", tmpls); err != nil {
		panic("failed loading templates: " + err.Error())
	}
	dynTrvlStyles = pgLocStyleFix(tmpls)
	endTrvlScrpit = pgEndScriptFix(tmpls)
	gx.MustIndexMap(&gxtTrvlFrame, needTemplate(tmpls, "topic"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtShipOpt, needTemplate(tmpls, "topic/shipopt"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtShipOptSel, needTemplate(tmpls, "topic/shipopt-sel"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtTrvlAvgHead, needTemplate(tmpls, "topic/avg-heading"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtTrvlAvgRow, needTemplate(tmpls, "topic/avg-row"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtTrvlAvgVal1, needTemplate(tmpls, "topic/avg-row/value1"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtTrvlAvgVal2, needTemplate(tmpls, "topic/avg-row/value2"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtTrvlDestRow, needTemplate(tmpls, "topic/dest"), idxMapNames.Convert)
}

var theAvgSteps = []int{5, 20, 50}

func emitAvgDurs(wr io.Writer,
	b1 *gx.BounT, ph []int,
	b2 *gx.BounT, sumPh, avgPh []int,
	data []time.Duration) (n int) {
	var bt *gx.BounT
	for i, dur := range data {
		if dur > 0 {
			if i == 0 {
				b1.BindP(ph, dur)
				bt = b1
			} else if i-1 < len(theAvgSteps) {
				avg := dur / time.Duration(theAvgSteps[i-1])
				b2.BindP(sumPh, dur)
				b2.BindP(avgPh, avg)
				bt = b2
			} else {
				avg := dur / time.Duration(len(theGame.JumpHist)-1)
				b2.BindP(sumPh, dur)
				b2.BindP(avgPh, avg.Round(time.Second))
				bt = b2
			}
		} else if i == 0 {
			b1.Bind(ph, webGuiNOC)
			bt = b1
		} else {
			b2.Bind(sumPh, webGuiNOC)
			b2.Bind(avgPh, webGuiNOC)
			bt = b2
		}
		n += bt.Emit(wr)
	}
	b2.Bind(sumPh, webGuiNOC)
	b2.Bind(avgPh, webGuiNOC)
	for i := len(data); i <= len(theAvgSteps); i++ {
		n += b2.Emit(wr)
	}
	return n
}

func emitAvgs(wr io.Writer,
	b1 *gx.BounT, ph []int,
	b2 *gx.BounT, sumPh, avgPh []int,
	data []float64) (n int) {
	var bt *gx.BounT
	for i, num := range data {
		if math.IsNaN(num) {
			b1.Bind(ph, webGuiNOC)
			bt = b1
		} else if i == 0 {
			b1.BindFmt(ph, "%.2f", num)
			bt = b1
		} else if i-1 < len(theAvgSteps) {
			avg := num / float64(theAvgSteps[i-1])
			b2.BindFmt(sumPh, "%.2f", num)
			b2.BindFmt(avgPh, "%.2f", avg)
			bt = b2
		} else {
			avg := num / float64(len(theGame.JumpHist)-1)
			b2.BindFmt(sumPh, "%.2f", num)
			b2.BindFmt(avgPh, "%.2f", avg)
			bt = b2
		}
		n += bt.Emit(wr)
	}
	b2.Bind(sumPh, webGuiNOC)
	b2.Bind(avgPh, webGuiNOC)
	for i := len(data); i <= len(theAvgSteps); i++ {
		n += b2.Emit(wr)
	}
	return n
}

func emitAvgRels(wr io.Writer, bt *gx.BounT, ph []int, d1, d2 []float64) (n int) {
	if len(d1) != len(d2) {
		panic("ral averages with inconsitent data")
	}
	for i := 0; i < len(d1); i++ {
		v1, v2 := d1[i], d2[i]
		if math.IsNaN(v1) || math.IsNaN(v2) {
			bt.Bind(ph, gx.Empty)
		} else if v2 == 0 {
			bt.BindP(ph, "∞")
		} else {
			q := v1 / v2
			sym := '→'
			switch {
			case q < 0.4:
				sym = '↺'
			case q < 0.7:
				sym = '↷'
			case q < 0.9:
				sym = '↝'
			}
			bt.BindFmt(ph, "%c %.2f%%", sym, 100.0*q)
		}
		n += bt.Emit(wr)
	}
	bt.Bind(ph, gx.Empty)
	for i := len(d2); i <= len(theAvgSteps); i++ {
		n += bt.Emit(wr)
	}
	return n
}

func emitAvgSpeeds(wr io.Writer, bt *gx.BounT, ph []int,
	lens []float64, durs []time.Duration) (n int) {
	for i := 0; i < len(lens); i++ {
		d := lens[i]
		t := durs[i]
		if t == 0 || math.IsNaN(d) {
			bt.Bind(ph, gx.Empty)
		} else {
			d /= t.Hours()
			bt.BindFmt(ph, "%.2f", d)
		}
		n += bt.Emit(wr)
	}
	bt.Bind(ph, gx.Empty)
	for i := len(lens); i <= len(theAvgSteps); i++ {
		n += bt.Emit(wr)
	}
	return n
}

func computeTravel() (dur []time.Duration, path []float64, dist []float64) {
	jhist := theGame.JumpHist
	if len(jhist) == 0 {
		return dur, path, dist
	}
	count, tCount := 0, 0
	var pathSum float64 = 0.0
	var dtSum time.Duration
	step := 0
	pos0 := jhist[len(jhist)-1]
	for i := len(jhist) - 1; i > 0; i-- {
		j0, j1 := jhist[i], jhist[i-1]
		pathSum += gxy.Dist(j0.Sys, j1.Sys)
		count++
		if !j0.First {
			dt := time.Time(j0.Arrive).Sub(time.Time(j1.Arrive))
			dtSum += dt
			tCount++
		}
		if count == 1 {
			path = append(path, pathSum)
			dist = append(dist, pathSum)
			if tCount == 1 {
				dur = append(dur, dtSum.Round(time.Second))
			} else {
				dur = append(dur, 0)
			}
		}
		if step < len(theAvgSteps) && count == theAvgSteps[step] {
			path = append(path, pathSum)
			dist = append(dist, gxy.Dist(pos0.Sys, j1.Sys))
			if tCount == count {
				dur = append(dur, dtSum.Round(time.Second))
			} else if tCount > 0 {
				avg := float64(dtSum) / float64(tCount)
				dur = append(dur, time.Duration(avg*float64(count)).Round(time.Second))
			} else {
				dur = append(dur, 0)
			}
			step++
		}
	}
	for len(dur) <= len(theAvgSteps) {
		path = append(path, math.NaN())
		dist = append(dist, math.NaN())
		dur = append(dur, 0)
	}
	path = append(path, pathSum)
	dist = append(dist, gxy.Dist(pos0.Sys, jhist[0].Sys))
	if tCount == count {
		dur = append(dur, dtSum.Round(time.Second))
	} else if tCount > 0 {
		avg := float64(dtSum) / float64(tCount)
		dur = append(dur, time.Duration(avg*float64(count)).Round(time.Second))
	} else {
		dur = append(dur, 0)
	}
	return dur, path, dist
}

func emitJumpStats(btFrame *gx.BounT, times []time.Duration, paths, dists []float64) {
	btFrame.BindGen(gxtTrvlFrame.AvgHeads, func(wr io.Writer) (n int) {
		btAvgHead := gxtTrvlAvgHead.NewBounT()
		for i := 0; i < len(theAvgSteps); i++ {
			btAvgHead.BindP(gxtTrvlAvgHead.Num, theAvgSteps[i])
			n += btAvgHead.Emit(wr)
		}
		btAvgHead.BindP(gxtTrvlAvgHead.Num, len(theGame.JumpHist)-1)
		n += btAvgHead.Emit(wr)
		return n
	})
	btAvgRow := gxtTrvlAvgRow.NewBounT()
	btAvgVal1 := gxtTrvlAvgVal1.NewBounT()
	btAvgVal2 := gxtTrvlAvgVal2.NewBounT()
	btFrame.BindGen(gxtTrvlFrame.AvgRows, func(wr io.Writer) (n int) {
		btAvgRow.BindP(gxtTrvlAvgRow.Title, "Distance / Path")
		btAvgRow.BindGen(gxtTrvlAvgRow.Vals, func(wr io.Writer) (n int) {
			return emitAvgRels(wr, btAvgVal1, gxtTrvlAvgVal1.Data, dists, paths)
		})
		n += btAvgRow.Emit(wr)
		btAvgRow.BindP(gxtTrvlAvgRow.Title, "Δt")
		btAvgRow.BindGen(gxtTrvlAvgRow.Vals, func(wr io.Writer) (n int) {
			return emitAvgDurs(wr,
				btAvgVal1, gxtTrvlAvgVal1.Data,
				btAvgVal2, gxtTrvlAvgVal2.Sum, gxtTrvlAvgVal2.Avg,
				times)
		})
		n += btAvgRow.Emit(wr)
		btAvgRow.BindP(gxtTrvlAvgRow.Title, "Distance [Ly]")
		btAvgRow.BindGen(gxtTrvlAvgRow.Vals, func(wr io.Writer) (n int) {
			return emitAvgs(wr,
				btAvgVal1, gxtTrvlAvgVal1.Data,
				btAvgVal2, gxtTrvlAvgVal2.Sum, gxtTrvlAvgVal2.Avg,
				dists)
		})
		n += btAvgRow.Emit(wr)
		btAvgRow.BindP(gxtTrvlAvgRow.Title, "Distance [Ly/h]")
		btAvgRow.BindGen(gxtTrvlAvgRow.Vals, func(wr io.Writer) (n int) {
			return emitAvgSpeeds(wr, btAvgVal1, gxtTrvlAvgVal1.Data, dists, times)
		})
		n += btAvgRow.Emit(wr)
		btAvgRow.BindP(gxtTrvlAvgRow.Title, "Path [Ly]")
		btAvgRow.BindGen(gxtTrvlAvgRow.Vals, func(wr io.Writer) (n int) {
			return emitAvgs(wr,
				btAvgVal1, gxtTrvlAvgVal1.Data,
				btAvgVal2, gxtTrvlAvgVal2.Sum, gxtTrvlAvgVal2.Avg,
				paths)
		})
		n += btAvgRow.Emit(wr)
		btAvgRow.BindP(gxtTrvlAvgRow.Title, "Path [Ly/h]")
		btAvgRow.BindGen(gxtTrvlAvgRow.Vals, func(wr io.Writer) (n int) {
			return emitAvgSpeeds(wr, btAvgVal1, gxtTrvlAvgVal1.Data, paths, times)
		})
		n += btAvgRow.Emit(wr)
		return n
	})
}

func emitDests(btFrame *gx.BounT, times []time.Duration, paths, dists []float64) {
	var dpjAvg, dpjMax float64 // distance per jump
	useJumpStats := 0.0        // ~ did we travel a straight line in jump history?
	jStatIdx := -1
	if len(dists) > 0 {
		const JS_LOWTH, JS_HIGTH = 0.8, 0.92
		jStatIdx = len(dists) - 1
		pathEffcy := dists[jStatIdx] / paths[jStatIdx]
		useJumpStats = (pathEffcy - JS_LOWTH) / (JS_HIGTH - JS_LOWTH)
	}
	if jStatIdx < 0 {
		dpjAvg = 0
	} else if useJumpStats <= 0.0 {
		dpjAvg = paths[jStatIdx] / float64(len(theGame.JumpHist)-1)
	} else if useJumpStats >= 1.0 {
		dpjAvg = dists[jStatIdx] / float64(len(theGame.JumpHist)-1)
	} else {
		dpjAvgP := paths[jStatIdx] / float64(len(theGame.JumpHist)-1)
		dpjAvgD := dists[jStatIdx] / float64(len(theGame.JumpHist)-1)
		dpjAvg = useJumpStats*dpjAvgD + (1-useJumpStats)*dpjAvgP
	}
	cmdr := &theGame.Cmdr
	planship := theGame.TrvlPlanShip.Ship
	if planship == nil {
		planship = cmdr.CurShip.Ship
		glog.Logf(l.Debug, "plannig with current ship")
	} else if planship != nil {
		glog.Logf(l.Debug, "plannig with ship %d %s / %s",
			planship.ID,
			planship.Name,
			planship.Ident)
	} else {
		glog.Logf(l.Debug, "no ship to plan with")
	}
	if planship == nil || planship.Jump.DistMax <= 0 {
		dpjMax = dpjAvg // same values → no interval
		glog.Logf(l.Debug, "no max jumprange for ship → use avg: %.2f", dpjAvg)
	} else {
		dpjMax = float64(planship.Jump.DistMax)
		if dpjAvg == 0 {
			dpjAvg = dpjMax // same values → no interval
			glog.Logf(l.Debug, "no avg jumprange → use ship max: %.2f", dpjMax)
		} else if dpjAvg > dpjMax {
			dpjAvg = dpjMax
			glog.Logf(l.Debug, "avg jumprange exceeds ships max: %f > %f", dpjAvg, dpjMax)
		}
	}
	btShipOpt := gxtShipOpt.NewBounT()
	btShipOptSel := gxtShipOptSel.NewBounT()
	btFrame.BindGen(gxtTrvlFrame.ShipOpts, func(wr io.Writer) (n int) {
		curship := cmdr.CurShip.Ship
		if curship != nil {
			if theGame.TrvlPlanShip.Ship == curship {
				btShipOptSel.BindP(gxtShipOptSel.Id, -1)
				btShipOptSel.BindP(gxtShipOptSel.Ship, "Current ship")
				btShipOptSel.BindFmt(gxtShipOptSel.Jump, "%.2f", curship.Jump.DistMax)
				n += btShipOptSel.Emit(wr)
			} else {
				btShipOpt.BindP(gxtShipOpt.Id, -1)
				btShipOpt.BindP(gxtShipOpt.Ship, "Current ship")
				btShipOpt.BindFmt(gxtShipOpt.Jump, "%.2f", curship.Jump.DistMax)
				n += btShipOpt.Emit(wr)
			}
		}
		for _, ship := range cmdr.Ships {
			if ship.ID < 0 || ship.Type == "testbuggy" {
				continue
			}
			shTy, _ := nmShipType.Map(ship.Type)
			if ship == theGame.TrvlPlanShip.Ship {
				btShipOptSel.BindP(gxtShipOptSel.Id, ship.ID)
				btShipOptSel.BindFmt(gxtShipOptSel.Ship, "%s: %s / %s",
					shTy,
					ship.Name,
					ship.Ident)
				btShipOptSel.BindFmt(gxtShipOptSel.Jump, "%.2f", ship.Jump.DistMax)
				n += btShipOptSel.Emit(wr)
			} else {
				btShipOpt.BindP(gxtShipOpt.Id, ship.ID)
				btShipOpt.BindFmt(gxtShipOpt.Ship, "%s: %s / %s",
					gxw.HtmlEsc(shTy),
					gxw.HtmlEsc(ship.Name),
					gxw.HtmlEsc(ship.Ident))
				btShipOpt.BindFmt(gxtShipOpt.Jump, "%.2f", ship.Jump.DistMax)
				n += btShipOpt.Emit(wr)
			}
		}
		return n
	})
	btFrame.BindFmt(gxtTrvlFrame.LypjMax, "%.2f", dpjMax)
	btFrame.BindFmt(gxtTrvlFrame.LypjAvg, "%.2f", dpjAvg)
	btDest := gxtTrvlDestRow.NewBounT()
	btFrame.BindGen(gxtTrvlFrame.Dests, func(wr io.Writer) (n int) {
		for i, dst := range cmdr.Dests {
			dstLoc := dst.Loc.Ref
			dist2 := gxy.Dist(dstLoc, cmdr.Loc.Ref)
			btDest.BindP(gxtTrvlDestRow.DstId, i)
			btDest.Bind(gxtTrvlDestRow.Name, CntLoc{dst.Loc.Ref})
			if cmdr.Home.Nil() || dst.Loc.Ref != cmdr.Home.Ref {
				btDest.BindP(gxtTrvlDestRow.HomeFlag, "not")
			} else {
				btDest.Bind(gxtTrvlDestRow.HomeFlag, gx.Empty)
			}
			btDest.Bind(gxtTrvlDestRow.Dist, gxm.Msg(wuiL7d, "%.2f", dist2))
			djAvg := int(math.Ceil(dist2 / dpjAvg))
			djMin := int(math.Ceil(dist2 / dpjMax))
			if djMin == djAvg {
				btDest.BindFmt(gxtTrvlDestRow.EJD, "~%d", djAvg)
			} else {
				btDest.BindFmt(gxtTrvlDestRow.EJD, "%d~%d", djMin, djAvg)
			}
			if theGame.Tj2j > 0 {
				dur2 := time.Duration(djAvg) * theGame.Tj2j
				btDest.BindP(gxtTrvlDestRow.ETD, dur2)
			} else {
				btDest.Bind(gxtTrvlDestRow.ETD, webGuiNOC)
			}
			btDest.Bind(gxtTrvlDestRow.CooX, gxm.Msg(wuiL7d, "%.2f", dstLoc.GCoos()[gxy.Xk]))
			btDest.Bind(gxtTrvlDestRow.CooY, gxm.Msg(wuiL7d, "%.2f", dstLoc.GCoos()[gxy.Yk]))
			btDest.Bind(gxtTrvlDestRow.CooZ, gxm.Msg(wuiL7d, "%.2f", dstLoc.GCoos()[gxy.Zk]))
			btDest.BindP(gxtTrvlDestRow.Note, gxw.HtmlEsc(dst.Note))
			btDest.BindP(gxtTrvlDestRow.Tags,
				gxw.HtmlEsc(str.Join(dst.Tags, ", ")))
			n += btDest.Emit(wr)
		}
		return n
	})
}

func wuiTravel(w http.ResponseWriter, r *http.Request) {
	btEmit, btBind, hook := preparePage(dynTrvlStyles, endTrvlScrpit, activeTopic(r))
	btFrame := gxtTrvlFrame.NewBounT()
	btBind.Bind(hook, btFrame)
	cmdr := &theGame.Cmdr
	if len(cmdr.Dests) == 0 && !cmdr.Home.Nil() { // TODO not very well placed here!!!
		home := &c.Destination{Loc: cmdr.Home}
		home.Tags = append(home.Tags, "Home")
		cmdr.Dests = append(cmdr.Dests, home)
	}
	times, paths, dists := computeTravel()
	emitJumpStats(btFrame, times, paths, dists)
	emitDests(btFrame, times, paths, dists)
	btEmit.Emit(w)
}

var trvlUsrOps = map[string]userHanlder{
	"planShip":  trvlPlanShip,
	"tglHomeId": trvlTglHmid,
	"addDst":    trvlAddDest,
	"delDst":    trvlDelDest,
}

func trvlPlanShip(gstat *c.GmState, evt map[string]interface{}) (reload bool) {
	jshid, ok := evt["shipId"]
	if ok {
		shid := int(jshid.(float64))
		var ship *c.Ship = nil
		if shid >= 0 {
			ship = gstat.Cmdr.ShipById(shid)
			if ship == nil {
				eulog.Logf(l.Warn, "cannot find ship with id %d", shid)
			}
		}
		reload = (gstat.TrvlPlanShip.Ship != ship)
		eulog.Logf(l.Trace, "plan ship: %v → %v => %t",
			gstat.TrvlPlanShip.Ship,
			ship,
			reload)
		gstat.TrvlPlanShip.Ship = ship
	} else {
		eulog.Logf(l.Error, "missing ship id in travel/plan-ship")
	}
	return reload
}

func trvlTglHmid(gstat *c.GmState, evt map[string]interface{}) (reload bool) {
	cmdr := &gstat.Cmdr
	dstIdx, _ := attInt(evt, "id")
	newHome := cmdr.Dests[dstIdx].Loc
	if newHome == cmdr.Home {
		cmdr.Home.Ref = nil
	} else {
		cmdr.Home = newHome
	}
	return true
}

func trvlAddDest(gstat *c.GmState, evt map[string]interface{}) (reload bool) {
	locStr, _ := attStr(evt, "nm")
	loc, _ := gxy.ParseLoc(locStr, theGalaxy)
	if loc == nil {
		// TODO no silent fail!
		return false
	} else {
		note, _ := attStr(evt, "note")
		tags, _ := attStr(evt, "tags")
		coos, _ := attStr(evt, "coo")
		cmdr := &gstat.Cmdr
		dst := cmdr.FindDest(loc)
		if dst == nil {
			dst = &c.Destination{
				Loc: c.LocRef{loc},
			}
			cmdr.Dests = append(cmdr.Dests, dst)
		}
		if len(coos) > 0 {
			ctxt := strings.Split(coos, "/")
			num := 3
			if len(ctxt) < num {
				num = len(ctxt)
			}
			sys := loc.System()
			for i := 0; i < num; i++ {
				glog.Logf(l.Trace, "parse dest coo %d = %s", i, ctxt[i])
				if f, err := parseDec(strings.TrimSpace(ctxt[i])); err == nil {
					sys.Coos[i] = f
				} else {
					sys.Coos[i] = math.NaN()
				}
			}
		}
		dst.Note = note
		dst.Tags = strings.Split(tags, ",")
		for i, tag := range dst.Tags {
			dst.Tags[i] = strings.TrimSpace(tag)
		}
		return true
	}
}

func trvlDelDest(gstat *c.GmState, evt map[string]interface{}) (reload bool) {
	cmdr := &gstat.Cmdr
	dstIdx, _ := attInt(evt, "id")
	if dstIdx >= 0 && dstIdx < len(cmdr.Dests) {
		loc := cmdr.Dests[dstIdx].Loc.Ref
		cmdr.RmDest(loc)
	}
	return true
}
