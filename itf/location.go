package itf

type TravelMode int

const (
	NoTMode TravelMode = iota
	Docked
	Cruise
	SCruise
	Jump
	Fly
	Drive

	TModeNum
)

var tmodeNames = [Drive]string{
	"docked",
	"cruise",
	"scruise",
	"jump",
	"fly",
	"drive",
}

func (tm TravelMode) String() string { return tmodeNames[tm-1] }

func ParseMode(s string) TravelMode {
	for i, n := range tmodeNames {
		if n == s {
			return TravelMode(i + 1)
		}
	}
	return NoTMode
}

type LocRefType int

const (
	NoRefType LocRefType = iota
	Star
	Planet
	Station

	LocRefNum
)

var reftypeNames = [Station]string{
	"star",
	"planet",
	"station",
}

func (lr LocRefType) String() string { return reftypeNames[lr-1] }

func ParseRefType(s string) LocRefType {
	for i, n := range reftypeNames {
		if n == s {
			return LocRefType(i + 1)
		}
	}
	return NoRefType
}

const (
	Lat = iota
	Lon
	Height
)

type LocInSys struct {
	Mode    TravelMode
	RefType LocRefType
	Ref     string
	Coos    []float64 `json:",omitempty"`
}

type Location struct {
	SysID   uint64
	SysName string
	LocInSys
}
