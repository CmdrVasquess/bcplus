package cmdr

import (
	"strings"
	"time"
)

const NoShip = -1

type Ship struct {
	Id          int
	Type        string
	Ident       string
	Name        string
	Bought      time.Time
	BerthLoc    int64
	Health      float64
	Rebuy       int
	HullValue   int
	ModuleValue int
}

func (s *State) MustHaveShip(id int, typ string) *Ship {
	if res, ok := s.Ships[id]; ok {
		if len(typ) > 0 && len(res.Type) == 0 {
			res.Type = typ
		}
		return res
	} else {
		res = &Ship{
			Id:   id,
			Type: strings.ToLower(typ),
		}
		s.Ships[id] = res
		return res
	}
}
