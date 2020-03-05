package itf

type TravelMode int

const (
	Unknown TravelMode = iota
	Docked
	Cruise
	SCruise
	Jump
	Fly
	Drive
)

type LocRefType int

const (
	Unknown LocRefType = iota
	Star
	Planet
	Station
)

type Location struct {
	SysID   uint64
	Mode    TravelMode
	RefType LocRefType
	Ref     string
	Coos    []float64 `json:",omitempty"`
}
