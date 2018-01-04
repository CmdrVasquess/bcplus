package main

import (
	"net/http"

	gx "git.fractalqb.de/goxic"
)

var _ = gx.Empty

func wuiDashboard(w http.ResponseWriter, r *http.Request) {
	btEmit, btBind, hook := preparePage(gx.Empty)
	btBind.BindP(hook,
		"…more content goes here! (If you see this anyway, choose a different topic from above)…")
	btEmit.Emit(w)
}
