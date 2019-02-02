package webui

import (
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/CmdrVasquess/BCplus/cmdr"
)

const (
	tkeyShips = "ships"
)

var gxtShips gxtTopic

type tpcShipData struct {
	Id                            int
	Type, Name, Ident             string
	Bought                        time.Time
	BSys, BLoc                    string
	Health                        float64
	Rebuy, HullValue, ModuleValue int
	Hardp, Util, Core, Opts       []*cmdr.Module
}

func TpcShipsData(c *cmdr.State) []*tpcShipData {
	res := make([]*tpcShipData, 0, len(c.Ships))
	for _, ship := range c.Ships {
		switch {
		case ship.Type == "testbuggy":
			continue
		case strings.HasSuffix(ship.Type, "_fighter"):
			continue
		}
		typ, _ := nmap.ShipType.Map(ship.Type)
		dto := &tpcShipData{
			Id:          ship.Id,
			Type:        typ,
			Name:        ship.Name,
			Ident:       ship.Ident,
			Bought:      ship.Bought,
			Rebuy:       ship.Rebuy,
			HullValue:   ship.HullValue,
			ModuleValue: ship.ModuleValue,
			Hardp:       ship.Hardpoints,
			Util:        ship.Utilities,
			Core:        ship.CoreModules,
			Opts:        ship.OptModules,
		}
		bloc, err := theGalaxy.GetSysPart(ship.BerthLoc)
		if err != nil {
			log.Errora("`err`", err)
		} else if bloc != nil {
			bsys, _ := theGalaxy.GetSystem(bloc.SysId) // TODO error
			dto.BSys = bsys.Name
			dto.BLoc = bloc.Name
		}
		res = append(res, dto)
	}
	sort.Slice(res, func(i, j int) bool {
		cmpr := strings.Compare(res[i].Type, res[j].Type)
		if cmpr == 0 {
			cmpr = strings.Compare(res[i].Name, res[j].Name)
		}
		return cmpr < 0
	})
	return res
}

func tpcShips(w http.ResponseWriter, r *http.Request) {
	bt := gxtShips.NewBounT(nil)
	bindTpcHeader(bt, &gxtShips)
	data := TpcShipsData(theCmdr())
	bt.BindGen(gxtShips.TopicData, jsonContent(&data))
	bt.Emit(w)
	currentTopic = UIShips
}
