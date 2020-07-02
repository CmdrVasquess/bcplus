package travel

import (
	"git.fractalqb.de/fractalqb/goxic"
	"github.com/CmdrVasquess/bcplus/internal/wapp"
)

func init() {
	wapp.AddScreen(&screen)
}

type template struct {
	wapp.Screen
	Data goxic.PhIdxs
}

var (
	tmpl   template
	screen = wapp.Screen{
		Key:      "travel",
		Tab:      "Travel",
		Title:    "Travel",
		Template: &tmpl,
	}
)
