package app

type InSysBody struct {
	Id              int
	Name            string
	Dist            float32
	R, Grav, Temp   float32
	Volcano         string
	Land, TidalLock bool
	Disco, Mapd     bool
}

type InSysInfo struct {
	BodyNum int
	MiscNum int
	Sigs    map[string]int
	Bodies  []*InSysBody
	BdyDsp  string `json:"bdyDsp"`
}

var inSysInfo = InSysInfo{
	BodyNum: -1,
	MiscNum: -1,
	Sigs:    make(map[string]int),
	BdyDsp:  "c",
}

func (isi *InSysInfo) reset() {
	isi.BodyNum = -1
	isi.MiscNum = -1
	isi.Sigs = make(map[string]int)
	isi.Bodies = nil
}

func (isi *InSysInfo) addSignal(nm string) {
	num := isi.Sigs[nm]
	isi.Sigs[nm] = num + 1
}
