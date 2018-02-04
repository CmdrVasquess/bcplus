package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"

	c "github.com/CmdrVasquess/BCplus/cmdr"
	gxy "github.com/CmdrVasquess/BCplus/galaxy"
	gx "github.com/fractalqb/goxic"
	gxm "github.com/fractalqb/goxic/textmessage"
	gxw "github.com/fractalqb/goxic/web"
	"github.com/fractalqb/namemap"
	"github.com/fractalqb/nmconv"
	l "github.com/fractalqb/qblog"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var idxMapNames = nmconv.Conversion{
	Norm:   nmconv.Uncamel,
	Denorm: nmconv.SepX(strings.ToLower, "-"),
}

func cmprMatByL7d(jnms []string, i, j int) bool {
	si := jnms[i]
	si, _ = nmMats.Map(si)
	sj := jnms[j]
	sj, _ = nmMats.Map(sj)
	return si < sj
}

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
	PgTitle     []int `goxic:"title"`
	Version     []int
	Styles      []int `goxic:"dyn-styles"`
	PgBody      []int `goxic:"body"`
	FullVersion []int
	ScriptEnd   []int
}

var gxtTitle struct {
	*gx.Template
	CmdrName []int
}

var gxtFrame struct {
	*gx.Template
	CmdrName   []int
	Credits    []int
	Loan       []int
	HomeFlag   []int
	DestFlag   []int
	RnkCombat  []int
	RnkTrade   []int
	RnkExplor  []int `goxic:"rnk-explorer"`
	RnkCqc     []int
	RLvlCombat []int `goxic:"rlvl-combat"`
	RLvlTrade  []int `goxic:"rlvl-trade"`
	RLvlExplor []int `goxic:"rlvl-explorer"`
	RLvlCqc    []int `goxic:"rlvl-cqc"`
	RPrgCombat []int `goxic:"rprg-combat"`
	RPrgTrade  []int `goxic:"rprg-trade"`
	RPrgExplor []int `goxic:"rprg-explorer"`
	RPrgCqc    []int `goxic:"rprg-cqc"`
	RnkFed     []int
	RLvlFed    []int `goxic:"rlvl-fed"`
	RPrgFed    []int `goxic:"rprg-fed"`
	RnkImp     []int
	RLvlImp    []int `goxic:"rlvl-imp"`
	RPrgImp    []int `goxic:"rprg-imp"`
	Loc        []int `goxic:"location"`
	LocX       []int `goxic:"locx"`
	LocY       []int `goxic:"locy"`
	LocZ       []int `goxic:"locz"`
	ShipType   []int
	ShipName   []int
	ShipIdent  []int
	Home       []int
	HomeDist   []int `goxic:"homedist"`
	NavItems   []int
	Topic      []int
}

var gxtNavItem struct {
	*gx.Template
	Link  []int
	Title []int
}

var gxtNavActv struct {
	*gx.Template
	Link  []int
	Title []int
}

func loadTmpls() {
	tmpls := make(map[string]*gx.Template)
	if err := gxw.ParseHtmlTemplate(assetPath("appframe.html"), "frame", tmpls); err != nil {
		panic("failed loading templates: " + err.Error())
	}
	gx.MustIndexMap(&gxtPage, needTemplate(tmpls, ""), idxMapNames.Convert)
	prepareOfflinePage(needTemplate(tmpls, "title-offline"),
		needTemplate(tmpls, "body-offline"))
	gx.MustIndexMap(&gxtTitle, needTemplate(tmpls, "title-online"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtFrame, needTemplate(tmpls, "body-online"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtNavItem, needTemplate(tmpls, "body-online/nav-item"), idxMapNames.Convert)
	gx.MustIndexMap(&gxtNavActv, needTemplate(tmpls, "body-online/nav-actv"), idxMapNames.Convert)
	loadWebTkTemplates()
	loadRescTemplates()
	loadTrvlTemplates()
	loadSynTemplates()
	loadShpTemplates()
	loadSysTemplates()
}

func prepareOfflinePage(title *gx.Template, body *gx.Template) {
	btOffline := gxtPage.NewBounT()
	if BCpBugfix == 0 {
		btOffline.BindFmt(gxtPage.Version, "%d.%d", BCpMajor, BCpMinor)
	} else {
		btOffline.BindFmt(gxtPage.Version, "%d.%d.%d", BCpMajor, BCpMinor, BCpBugfix)
	}
	btOffline.BindP(gxtPage.FullVersion, gxw.HtmlEsc(BCpDescStr()))
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
	btOffline.Bind(gxtPage.ScriptEnd, gx.Empty)
	fix := btOffline.Fixate()
	if stat, ok := fix.Static(); ok {
		offlinePage = stat
	} else {
		panic("offline page not static")
	}
}

func init() {
	loadTmpls()
}

func offline(w http.ResponseWriter, r *http.Request, h func(http.ResponseWriter, *http.Request)) {
	if theGame.IsOffline() {
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
	return gxw.EscHtml{gx.Print{str}}
}

func nmapU8(nm *namemap.FromTo, rank uint8) gx.Content {
	str, _ := nm.Map(strconv.Itoa(int(rank)))
	return gxw.EscHtml{gx.Print{str}}
}

func pgLocStyleFix(tmpls map[string]*gx.Template) (res gx.Content) {
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

func pgEndScriptFix(tmpls map[string]*gx.Template) (res gx.Content) {
	if escr, ok := tmpls["end-script"]; ok {
		if raw, ok := escr.Static(); ok {
			res = gx.Data(raw)
		} else {
			res = gx.Empty
		}
	} else {
		res = gx.Empty
	}
	return res
}

func pgEndScript(tmpls map[string]*gx.Template) *gx.Template {
	escr, _ := tmpls["end-script"]
	return escr
}

func emitNavItems(wr io.Writer, active string) (n int) {
	btNavi := gxtNavItem.NewBounT()
	btNava := gxtNavActv.NewBounT()
	var bt *gx.BounT
	for _, ln := range webGuiTopics {
		if ln == active {
			bt = btNava
		} else {
			bt = btNavi
		}
		bt.BindP(gxtNavItem.Link, ln)
		bt.Bind(gxtNavItem.Title, nmap(&nmNavItem, ln))
		n += bt.Emit(wr)
	}
	return n
}

func preparePage(styles, endScript gx.Content, active string) (emit, bindto *gx.BounT, hook []int) {
	cmdr := &theGame.Cmdr
	cmdrNameEsc := gxw.HtmlEsc(cmdr.Name)
	btPage := gxtPage.NewBounT()

	btTitle := gxtTitle.NewBounT()
	if BCpBugfix == 0 {
		btPage.BindFmt(gxtPage.Version, "%d.%d", BCpMajor, BCpMinor)
	} else {
		btPage.BindFmt(gxtPage.Version, "%d.%d.%d", BCpMajor, BCpMinor, BCpBugfix)
	}
	btPage.BindP(gxtPage.FullVersion, gxw.HtmlEsc(BCpDescStr()))
	btPage.Bind(gxtPage.PgTitle, gxw.EscHtml{btTitle})
	btPage.Bind(gxtPage.Styles, styles)
	btTitle.BindP(gxtTitle.CmdrName, cmdrNameEsc)

	//	btFrame := gxtFrame.NewInitBounT(gx.Empty)
	btFrame := gxtFrame.NewBounT()
	btPage.Bind(gxtPage.PgBody, btFrame)
	btPage.Bind(gxtPage.ScriptEnd, endScript)
	btFrame.BindP(gxtFrame.CmdrName, cmdrNameEsc)
	// TODO "golang.org/x/text/message"
	btFrame.Bind(gxtFrame.Credits, gxm.Msg(wuiL7d, "%d", cmdr.Credits))
	btFrame.Bind(gxtFrame.Loan, gxm.Msg(wuiL7d, "%d", cmdr.Loan))
	if cmdr.Loc == cmdr.Home {
		btFrame.Bind(gxtFrame.HomeFlag, gx.Empty)
	} else {
		btFrame.BindP(gxtFrame.HomeFlag, "not")
	}
	if cmdr.FindDest(cmdr.Loc.Ref) != nil {
		btFrame.Bind(gxtFrame.DestFlag, gx.Empty)
	} else {
		btFrame.BindP(gxtFrame.DestFlag, "no")
	}

	btFrame.Bind(gxtFrame.RnkCombat, nmapU8(&nmRnkCombat, cmdr.Ranks[c.RnkCombat]))
	btFrame.BindP(gxtFrame.RLvlCombat, cmdr.Ranks[c.RnkCombat])
	btFrame.BindP(gxtFrame.RPrgCombat, cmdr.RnkPrgs[c.RnkCombat])
	btFrame.Bind(gxtFrame.RnkTrade, nmapU8(&nmRnkTrade, cmdr.Ranks[c.RnkTrade]))
	btFrame.BindP(gxtFrame.RLvlTrade, cmdr.Ranks[c.RnkTrade])
	btFrame.BindP(gxtFrame.RPrgTrade, cmdr.RnkPrgs[c.RnkTrade])
	btFrame.Bind(gxtFrame.RnkExplor, nmapU8(&nmRnkExplor, cmdr.Ranks[c.RnkExplore]))
	btFrame.BindP(gxtFrame.RLvlExplor, cmdr.Ranks[c.RnkExplore])
	btFrame.BindP(gxtFrame.RPrgExplor, cmdr.RnkPrgs[c.RnkExplore])
	btFrame.Bind(gxtFrame.RnkCqc, nmapU8(&nmRnkCqc, cmdr.Ranks[c.RnkCqc]))
	btFrame.BindP(gxtFrame.RLvlCqc, cmdr.Ranks[c.RnkCqc])
	btFrame.BindP(gxtFrame.RPrgCqc, cmdr.RnkPrgs[c.RnkCqc])
	btFrame.Bind(gxtFrame.RnkFed, nmapU8(&nmRnkFed, cmdr.Ranks[c.RnkFed]))
	btFrame.BindP(gxtFrame.RLvlFed, cmdr.Ranks[c.RnkFed])
	btFrame.BindP(gxtFrame.RPrgFed, cmdr.RnkPrgs[c.RnkFed])
	btFrame.Bind(gxtFrame.RnkImp, nmapU8(&nmRnkImp, cmdr.Ranks[c.RnkImp]))
	btFrame.BindP(gxtFrame.RLvlImp, cmdr.Ranks[c.RnkImp])
	btFrame.BindP(gxtFrame.RPrgImp, cmdr.RnkPrgs[c.RnkImp])
	if cmdr.Loc.Ref == nil {
		btFrame.Bind(gxtFrame.Loc, webGuiNOC)
		btFrame.Bind(gxtFrame.LocX, webGuiNOC)
		btFrame.Bind(gxtFrame.LocY, webGuiNOC)
		btFrame.Bind(gxtFrame.LocZ, webGuiNOC)
	} else {
		btFrame.Bind(gxtFrame.Loc, gxw.EscHtml{gx.Print{cmdr.Loc.String()}})
		sysCoos := cmdr.Loc.Ref.GCoos()
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
		if len(cshp.Name) == 0 {
			btFrame.Bind(gxtFrame.ShipName, webGuiNOC)
		} else {
			btFrame.Bind(gxtFrame.ShipName, gxw.EscHtml{gx.Print{cshp.Name}})
		}
		if len(cshp.Ident) == 0 {
			btFrame.Bind(gxtFrame.ShipIdent, webGuiNOC)
		} else {
			btFrame.Bind(gxtFrame.ShipIdent, gxw.EscHtml{gx.Print{cshp.Ident}})
		}
	}
	if cmdr.Home.Ref == nil {
		btFrame.Bind(gxtFrame.Home, webGuiNOC)
		btFrame.Bind(gxtFrame.HomeDist, webGuiNOC)
	} else {
		btFrame.Bind(gxtFrame.Home, gxw.EscHtml{gx.Print{cmdr.Home.String()}})
		btFrame.Bind(gxtFrame.HomeDist,
			gxm.Msg(wuiL7d, "%.2f", gxy.Dist(cmdr.Home.Ref, cmdr.Loc.Ref)))
	}
	btFrame.BindGen(gxtFrame.NavItems, func(wr io.Writer) int {
		return emitNavItems(wr, active)
	})

	return btPage, btFrame, gxtFrame.Topic
}

func setupTopic(link string, handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc("/"+link, func(w http.ResponseWriter, r *http.Request) {
		offline(w, r, handler)
	})
	webGuiTopics = append(webGuiTopics, link)
}

func activeTopic(r *http.Request) (res string) {
	res = r.URL.Path[1:]
	glog.Logf(l.Trace, "web ui: active topic '%s'", res)
	return res
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
	setupTopic("ships", wuiShp)
	setupTopic("travel", wuiTravel)
	setupTopic("system", wuiSys)
	setupTopic("materials", wuiMats)
	setupTopic("synth", wuiSyn)
	glog.Logf(l.Info, "Starting web GUI on port %d", webGuiPort)
	go http.ListenAndServe(fmt.Sprintf(":%d", webGuiPort), nil)
	ifaddrs, _ := net.InterfaceAddrs()
	for _, addr := range ifaddrs {
		if nip, ok := addr.(*net.IPNet); ok && !nip.IP.IsLoopback() && nip.IP.To4() != nil {
			glog.Logf(l.Info,
				"for web GUI open 'http://%s:%d/' in your browser",
				nip.IP.String(),
				webGuiPort)
			break
		}
	}
}
