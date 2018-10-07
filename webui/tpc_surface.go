package webui

import (
	"net/http"

	"github.com/CmdrVasquess/BCplus/cmdr"
)

const (
	tkeySurface = "surface"
)

var gxtSurface gxtTopic

type tpcSurfaceData struct {
	Alt, Lat, Lon, Head cmdr.CooNaN
	Surf                *cmdr.Surface
}

func TpcSurfaceData(c *cmdr.State) *tpcSurfaceData {
	return &tpcSurfaceData{
		Alt:  c.Loc.Alt,
		Lat:  c.Loc.Lat,
		Lon:  c.Loc.Lon,
		Head: c.Loc.Heading,
		Surf: &c.Surface,
	}
}

func tpcSurface(w http.ResponseWriter, r *http.Request) {
	bt := gxtSurface.NewBounT(nil)
	bindTpcHeader(bt, &gxtSurface)
	data := TpcSurfaceData(theCmdr())
	bt.BindGen(gxtSurface.TopicData, jsonContent(data))
	bt.Emit(w)
	CurrentTopic = UISurface
}
