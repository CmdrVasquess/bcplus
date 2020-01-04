package app

type PosRef int

const (
	RefUndef PosRef = iota
	Space
	Star
	Belt
	Planet
	Ring
	Station
	Outpost
	Port
	Settlement
	JTarget
)

type Mode int

const (
	ModeUndef Mode = iota
	Parked
	Move
	Cruise
	Jump
)

type Vehicle int

const (
	VhclUndef Vehicle = iota
	InShip
	InSRV
	InFighter
	AsCrew
)

type SurfPos struct {
	LatLon [2]float64 `json:"ll"`
	Alt    float64    `json:"a"`
}

func (p *SurfPos) SetLatLon(lat, lon float64) (chg Change) {
	if p.LatLon[0] != lat {
		p.LatLon[0] = lat
		chg = ChgPos
	}
	if p.LatLon[1] != lon {
		p.LatLon[1] = lon
		chg |= ChgPos
	}
	return chg
}

func (p *SurfPos) SetAlt(a float64) (chg Change) {
	if p.Alt != a {
		p.Alt = a
		return ChgPos
	}
	return 0
}

type Location struct {
	SysId uint64   `json:"si"`
	SysNm string   `json:"sn"`
	Ref   PosRef   `json:"r"`
	RefNm string   `json:"rn"`
	Vhcl  Vehicle  `json:"v"`
	Mode  Mode     `json:"m"`
	Surf  *SurfPos `json:"sc,omitempty"`
}

func (l *Location) SetSys(id uint64, nm string) (chg Change) {
	if l.SysId == id {
		return 0
	}
	l.SysId = id
	l.SysNm = nm
	return ChgLoc
}

func (l *Location) SetRef(r PosRef) (chg Change) {
	if r != l.Ref {
		chg = ChgLoc
		l.Ref = r
	}
	return chg
}

func (l *Location) SetRefNm(nm string) (chg Change) {
	if nm != l.RefNm {
		chg = ChgLoc
		l.RefNm = nm
	}
	return chg
}

func (l *Location) SetVehicle(v Vehicle) (chg Change) {
	if v != l.Vhcl {
		l.Vhcl = v
		chg = ChgLoc
	}
	return chg
}

func (l *Location) SetMode(m Mode) (chg Change) {
	if m != l.Mode {
		l.Mode = m
		chg = ChgLoc
	}
	return chg
}

func (l *Location) SetSurf(p *SurfPos) (chg Change) {
	if l.Surf == nil {
		if p == nil {
			return 0
		}
		l.Surf = p
		return ChgPos
	} else if p == nil {
		l.Surf = nil
		return ChgPos
	} else if *l.Surf != *p {
		chg = ChgPos
	}
	l.Surf = p
	return chg
}
