package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	gxy "github.com/CmdrVasquess/BCplus/galaxy"
	gx "github.com/fractalqb/goxic"
	gxm "github.com/fractalqb/goxic/textmessage"
	gxw "github.com/fractalqb/goxic/web"
	"github.com/fractalqb/namemap"
	l "github.com/fractalqb/qblog"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func needTemplate(tmap map[string]*gx.Template, path string) *gx.Template {
	if t, ok := tmap[path]; !ok {
		glog.Fatalf("missing template: '%s'", path)
		return nil
	} else {
		return t
	}
}

var offlinePage []byte

var gxtPage struct {
	*gx.Template
	PgTitle []int `goxic:"title"`
	Styles  []int `goxic:"dyn-styles"`
	PgBody  []int `goxic:"body"`
}

var gxtTitle struct {
	*gx.Template
	CmdrName []int `goxic:"cmdr-name"`
}

var gxtFrame struct {
	*gx.Template
	CmdrName   []int `goxic:"cmdr-name"`
	Credits    []int `goxic:"credits"`
	Loan       []int `goxic:"loan"`
	RnkCombat  []int `goxic:"rnk-combat"`
	RnkTrade   []int `goxic:"rnk-trade"`
	RnkExplor  []int `goxic:"rnk-explorer"`
	RnkCqc     []int `goxic:"rnk-cqc"`
	RLvlCombat []int `goxic:"rlvl-combat"`
	RLvlTrade  []int `goxic:"rlvl-trade"`
	RLvlExplor []int `goxic:"rlvl-explorer"`
	RLvlCqc    []int `goxic:"rlvl-cqc"`
	RPrgCombat []int `goxic:"rprg-combat"`
	RPrgTrade  []int `goxic:"rprg-trade"`
	RPrgExplor []int `goxic:"rprg-explorer"`
	RPrgCqc    []int `goxic:"rprg-cqc"`
	RnkFed     []int `goxic:"rnk-fed"`
	RLvlFed    []int `goxic:"rlvl-fed"`
	RPrgFed    []int `goxic:"rprg-fed"`
	RnkImp     []int `goxic:"rnk-imp"`
	RLvlImp    []int `goxic:"rlvl-imp"`
	RPrgImp    []int `goxic:"rprg-imp"`
	Loc        []int `goxic:"location"`
	LocX       []int `goxic:"locx"`
	LocY       []int `goxic:"locy"`
	LocZ       []int `goxic:"locz"`
	ShipType   []int `goxic:"ship-type"`
	ShipName   []int `goxic:"ship-name"`
	ShipIdent  []int `goxic:"ship-ident"`
	Home       []int `goxic:"home"`
	HomeDist   []int `goxic:"homedist"`
	NavItems   []int `goxic:"nav-items"`
	Topic      []int `goxic:"topic"`
}

var gxtNavItem struct {
	*gx.Template
	Link  []int `goxic:"link"`
	Title []int `goxic:"title"`
}

func loadTmpls() {
	tmpls := make(map[string]*gx.Template)
	if err := gxw.ParseHtmlTemplate(assetPath("appframe.html"), "frame", tmpls); err != nil {
		panic("failed loading templates: " + err.Error())
	}
	gx.MustIndexMap(&gxtPage, needTemplate(tmpls, ""))
	prepareOfflinePage(needTemplate(tmpls, "title-offline"),
		needTemplate(tmpls, "body-offline"))
	gx.MustIndexMap(&gxtTitle, needTemplate(tmpls, "title-online"))
	gx.MustIndexMap(&gxtFrame, needTemplate(tmpls, "body-online"))
	gx.MustIndexMap(&gxtNavItem, needTemplate(tmpls, "body-online/nav-item"))
	loadRescTemplates()
	loadTrvlTemplates()
}

func prepareOfflinePage(title *gx.Template, body *gx.Template) {
	btOffline := gxtPage.NewBounT()
	if stat, ok := title.Static(); !ok {
		glog.Fatal("no offline title")
	} else {
		btOffline.BindP(gxtPage.PgTitle, string(stat))
	}
	if stat, ok := body.Static(); ok {
		btOffline.BindP(gxtPage.PgBody, string(stat))
	} else {
		glog.Fatal("no offline body")
	}
	btOffline.Bind(gxtPage.Styles, gx.Empty)
	if fix, err := btOffline.Fixate(); err != nil {
		panic("cannot fixate offline page: " + err.Error())
	} else if stat, ok := fix.Static(); ok {
		offlinePage = stat
	} else {
		panic("offline page not static: " + err.Error())
	}
}

func init() {
	loadTmpls()
}

func offline(w http.ResponseWriter, r *http.Request, h func(http.ResponseWriter, *http.Request)) {
	if theGame.isOffline() {
		w.Write(offlinePage)
	} else {
		theStateLock.RLock()
		defer theStateLock.RUnlock()
		h(w, r)
	}
}

var webGuiTBD = gx.Print{"???"}
var webGuiNOC = gx.Print{"–"}
var webGuiPort uint
var webGuiTopics []string
var wuiL7d = message.NewPrinter(language.Make("en"))

func nmap(nm *namemap.FromTo, term string) gx.Content {
	str, _ := nm.Map(term)
	return gx.Print{str}
}

func nmapU8(nm *namemap.FromTo, rank uint8) gx.Content {
	str, _ := nm.Map(strconv.Itoa(int(rank)))
	return gx.Print{str}
}

func pageLocalStyle(tmpls map[string]*gx.Template) (res gx.Content) {
	if lsty, ok := tmpls["local-style"]; ok {
		if raw, ok := lsty.Static(); ok {
			res = gx.Data(raw)
		} else {
			res = gx.Empty
		}
	} else {
		res = gx.Empty
	}
	return res
}

func emitNavItems(wr io.Writer) (n int) {
	btNavi := gxtNavItem.NewBounT()
	for _, ln := range webGuiTopics {
		btNavi.BindP(gxtNavItem.Link, ln)
		btNavi.Bind(gxtNavItem.Title, gxw.EscHtml{nmap(&nmNavItem, ln)})
		n += btNavi.Emit(wr)
	}
	return n
}

func preparePage(styles gx.Content) (emit, bindto *gx.BounT, hook []int) {
	cmdr := &theGame.Cmdr
	cmdrNameEsc := gxw.HtmlEsc(cmdr.Name)
	btPage := gxtPage.NewBounT()

	btTitle := gxtTitle.NewBounT()
	btPage.Bind(gxtPage.PgTitle, btTitle)
	btPage.Bind(gxtPage.Styles, styles)
	btTitle.BindP(gxtTitle.CmdrName, cmdrNameEsc)

	//	btFrame := gxtFrame.NewInitBounT(gx.Empty)
	btFrame := gxtFrame.NewBounT()
	btPage.Bind(gxtPage.PgBody, btFrame)
	btFrame.BindP(gxtFrame.CmdrName, cmdrNameEsc)
	// TODO "golang.org/x/text/message"
	btFrame.Bind(gxtFrame.Credits, gxw.EscHtml{gxm.Msg(wuiL7d, "%d", cmdr.Credits)})
	btFrame.Bind(gxtFrame.Loan, gxw.EscHtml{gxm.Msg(wuiL7d, "%d", cmdr.Loan)})

	btFrame.Bind(gxtFrame.RnkCombat, nmapU8(&nmRnkCombat, cmdr.Ranks[RnkCombat]))
	btFrame.BindP(gxtFrame.RLvlCombat, cmdr.Ranks[RnkCombat])
	btFrame.BindP(gxtFrame.RPrgCombat, cmdr.RnkPrgs[RnkCombat])
	btFrame.Bind(gxtFrame.RnkTrade, nmapU8(&nmRnkTrade, cmdr.Ranks[RnkTrade]))
	btFrame.BindP(gxtFrame.RLvlTrade, cmdr.Ranks[RnkTrade])
	btFrame.BindP(gxtFrame.RPrgTrade, cmdr.RnkPrgs[RnkTrade])
	btFrame.Bind(gxtFrame.RnkExplor, nmapU8(&nmRnkExplor, cmdr.Ranks[RnkExplore]))
	btFrame.BindP(gxtFrame.RLvlExplor, cmdr.Ranks[RnkExplore])
	btFrame.BindP(gxtFrame.RPrgExplor, cmdr.RnkPrgs[RnkExplore])
	btFrame.Bind(gxtFrame.RnkCqc, nmapU8(&nmRnkCqc, cmdr.Ranks[RnkCqc]))
	btFrame.BindP(gxtFrame.RLvlCqc, cmdr.Ranks[RnkCqc])
	btFrame.BindP(gxtFrame.RPrgCqc, cmdr.RnkPrgs[RnkCqc])
	btFrame.Bind(gxtFrame.RnkFed, nmapU8(&nmRnkFed, cmdr.Ranks[RnkFed]))
	btFrame.BindP(gxtFrame.RLvlFed, cmdr.Ranks[RnkFed])
	btFrame.BindP(gxtFrame.RPrgFed, cmdr.RnkPrgs[RnkFed])
	btFrame.Bind(gxtFrame.RnkImp, nmapU8(&nmRnkImp, cmdr.Ranks[RnkImp]))
	btFrame.BindP(gxtFrame.RLvlImp, cmdr.Ranks[RnkImp])
	btFrame.BindP(gxtFrame.RPrgImp, cmdr.RnkPrgs[RnkImp])
	if cmdr.Loc.Location == nil {
		btFrame.Bind(gxtFrame.Loc, webGuiNOC)
		btFrame.Bind(gxtFrame.LocX, webGuiNOC)
		btFrame.Bind(gxtFrame.LocY, webGuiNOC)
		btFrame.Bind(gxtFrame.LocZ, webGuiNOC)
	} else {
		btFrame.Bind(gxtFrame.Loc, gxw.EscHtml{gx.Print{cmdr.Loc.String()}})
		sysCoos := cmdr.Loc.GCoos()
		btFrame.Bind(gxtFrame.LocX, gxm.Msg(wuiL7d, "%.2f", sysCoos[gxy.Xk]))
		btFrame.Bind(gxtFrame.LocY, gxm.Msg(wuiL7d, "%.2f", sysCoos[gxy.Yk]))
		btFrame.Bind(gxtFrame.LocZ, gxm.Msg(wuiL7d, "%.2f", sysCoos[gxy.Zk]))
	}
	if cshp := cmdr.CurShip.Ship; cshp == nil {
		btFrame.Bind(gxtFrame.ShipType, webGuiNOC)
		btFrame.Bind(gxtFrame.ShipName, webGuiNOC)
		btFrame.Bind(gxtFrame.ShipIdent, webGuiNOC)
	} else {
		btFrame.Bind(gxtFrame.ShipType, nmap(&nmShipType, cshp.Type))
		btFrame.Bind(gxtFrame.ShipName, gxw.EscHtml{gx.Print{cshp.Name}})
		btFrame.Bind(gxtFrame.ShipIdent, gxw.EscHtml{gx.Print{cshp.Ident}})
	}
	if cmdr.Home.Location == nil {
		btFrame.Bind(gxtFrame.Home, webGuiNOC)
		btFrame.Bind(gxtFrame.HomeDist, webGuiNOC)
	} else {
		btFrame.Bind(gxtFrame.Home, gxw.EscHtml{gx.Print{cmdr.Home.String()}})
		btFrame.Bind(gxtFrame.HomeDist,
			gxm.Msg(wuiL7d, "%.2f", gxy.Dist(cmdr.Home, cmdr.Loc)))
	}
	btFrame.BindGen(gxtFrame.NavItems, emitNavItems)

	return btPage, btFrame, gxtFrame.Topic
}

func setupTopic(link string, handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc("/"+link, func(w http.ResponseWriter, r *http.Request) {
		offline(w, r, handler)
	})
	webGuiTopics = append(webGuiTopics, link)
}

func runWebGui() {
	htfs := http.FileServer(http.Dir(assetPath("s")))
	http.Handle("/s/", http.StripPrefix("/s", htfs))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		offline(w, r, wuiDashboard)
	})
	go wscHub()
	http.HandleFunc("/ws", serveWs)
	setupTopic("dashboard", wuiDashboard)
	setupTopic("resources", wuiResources)
	setupTopic("travel", wuiTravel)
	glog.Logf(l.Info, "Starting web GUI on port %d", webGuiPort)
	go http.ListenAndServe(fmt.Sprintf(":%d", webGuiPort), nil)
}
