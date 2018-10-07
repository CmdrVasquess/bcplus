package webui

import (
	"net/http"
	"sort"
	"strings"

	"github.com/CmdrVasquess/BCplus/cmdr"
)

const (
	tkeyShips = "ships"
)

var gxtShips gxtTopic

type tpcShipData struct {
	Id                int
	Type, Name, Ident string
	BSys, BLoc        string
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
			Id:    ship.Id,
			Type:  typ,
			Name:  ship.Name,
			Ident: ship.Ident,
		}
		bloc, err := theGalaxy.GetSysPart(ship.BerthLoc)
		if err != nil {
			log.Error(err)
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
	bt.BindGen(gxtShips.TopicData, jsonContent(data))
	bt.Emit(w)
	CurrentTopic = UIShips
}
