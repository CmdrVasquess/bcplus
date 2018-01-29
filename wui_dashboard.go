package main

import (
	"net/http"

	gx "github.com/fractalqb/goxic"
)

var _ = gx.Empty

func wuiDashboard(w http.ResponseWriter, r *http.Request) {
	btEmit, btBind, hook := preparePage(gx.Empty, gx.Empty, activeTopic(r))
	btBind.BindP(hook,
		`…more content goes here! (If you see this anyway, choose a different topic from above)…
		<br><p>BC+ is still in an early stage. Hopefully it will be useful to you in some way.</p>`)
	btEmit.Emit(w)
}
