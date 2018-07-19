package main

import (
	"io"
	"net/http"

	gxy "github.com/CmdrVasquess/BCplus/galaxy"
	gx "git.fractalqb.de/fractalqb/goxic"
	gxm "git.fractalqb.de/fractalqb/goxic/textmessage"
	gxw "git.fractalqb.de/fractalqb/goxic/web"
)

var gxtShpFrame struct {
	*gx.Template
	CurShips  []int
	SoldShips []int
}

var gxtCShip struct {
	*gx.Template
	Type  []int
	Name  []int
	Ident []int
	Jump  []int
	Loc   []int
	Dist  []int
}

var gxtSShip struct {
	*gx.Template
	Type   []int
	Name   []int
	Ident  []int
	Bought []int
	Sold   []int
}

func loadShpTemplates() {
	tmpls := make(map[string]*gx.Template)
	tpars := gxw.NewHtmlParser()
	if err := tpars.ParseFile(assetPath("ships.html"), "ships", tmpls); err != nil {
		panic("failed loading templates: " + err.Error())
	}
	//	dynShpStyles = pgLocStyleFix(tmpls)
	//	endShpScrpit = pgEndScriptFix(tmpls)
	gx.MustIndexMap(&gxtShpFrame, needTemplate(tmpls, "topic"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtCShip, needTemplate(tmpls, "topic/cur-ship"), idxMapNames.Convert)
}

func wuiShp(w http.ResponseWriter, r *http.Request) {
	btEmit, btBind, hook := preparePage(gx.Empty, gx.Empty, gx.Empty, activeTopic(r))
	btFrame := gxtShpFrame.NewBounT(nil)
	btBind.Bind(hook, btFrame)
	btCShip := gxtCShip.NewBounT(nil)
	btFrame.Bind(gxtShpFrame.CurShips, btCShip)
	cmdr := &theGame.Cmdr
	btFrame.BindGen(gxtShpFrame.CurShips, func(wr io.Writer) (n int) {
		for _, ship := range cmdr.Ships {
			kind, _ := nmShipType.MapNm(ship.Type, "kind")
			if ship.Sold != nil || kind != "S" {
				continue
			}
			shTy, _ := nmShipType.MapNm(ship.Type, "lang:")
			btCShip.BindP(gxtCShip.Ident, gxw.EscHtml(ship.Ident))
			btCShip.BindP(gxtCShip.Name, gxw.EscHtml(ship.Name))
			btCShip.BindP(gxtCShip.Type, gxw.EscHtml(shTy))
			btCShip.BindFmt(gxtCShip.Jump, "%.2f", ship.Jump.DistMax)
			if ship.Loc.Nil() {
				btCShip.Bind(gxtCShip.Loc, webGuiNOC)
				btCShip.Bind(gxtCShip.Dist, webGuiNOC)
			} else {
				btCShip.Bind(gxtCShip.Loc, CntLoc{ship.Loc.Ref}) //gxw.EscHtml(ship.Loc.String()))
				btCShip.Bind(gxtCShip.Dist,
					gxm.Msg(wuiL7d, "%.2f", gxy.Dist(ship.Loc.Ref, cmdr.Loc.Ref)))
			}
			n += btCShip.Emit(wr)
		}
		return n
	})
	btFrame.Bind(gxtShpFrame.SoldShips, gx.Empty)
	btEmit.Emit(w)
}
