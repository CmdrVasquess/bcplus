package main

import (
	"net/http"

	gx "github.com/fractalqb/goxic"
	gxw "github.com/fractalqb/goxic/web"
)

var gxcDshb gx.Content

func loadDshbTemplates() {
	tmpls := make(map[string]*gx.Template)
	if err := gxw.ParseHtmlTemplate(assetPath("dashboard.html"), "dshb", tmpls); err != nil {
		panic("failed loading templates: " + err.Error())
	}
	//	dynShpStyles = pgLocStyleFix(tmpls)
	//	endShpScrpit = pgEndScriptFix(tmpls)
	raw, ok := tmpls["topic"].Static()
	if !ok {
		glog.Fatal("dashboard template is not static content")
	}
	gxcDshb = gx.Data(raw)
}

func wuiDashboard(w http.ResponseWriter, r *http.Request) {
	btEmit, btBind, hook := preparePage(gx.Empty, gx.Empty, activeTopic(r))
	btBind.Bind(hook, gxcDshb)
	btEmit.Emit(w)
}
