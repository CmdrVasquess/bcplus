package cmdr

import (
	"fmt"
	"strconv"
	"strings"
)

type MatKind int

//go:generate stringer -type MatKind
const (
	MatRaw MatKind = iota
	MatMan
	MatEnc
)

// Material must be usable as key in a map
type Material string

func MatDefine(kind MatKind, key string) Material {
	return Material(fmt.Sprintf("%d:%s", kind, key))
}

func (mat Material) Parse() (kind MatKind, key string) {
	tmpKind, tmpKey := defParse(string(mat))
	return MatKind(tmpKind), tmpKey
}

type MatState struct {
	Have int
	Want int
}

type RcpKind int

//go:generate stringer -type RcpKind
const (
	RcpSynth RcpKind = iota
	RcpEngie
)

// RcpDef must be usable as key in a map
type RcpDef string

func RcpDefine(kind RcpKind, key string) RcpDef {
	return RcpDef(fmt.Sprintf("%d:%s", kind, key))
}

func (rcp RcpDef) Parse() (kind RcpKind, key string) {
	tmpKind, tmpKey := defParse(string(rcp))
	return RcpKind(tmpKind), tmpKey
}

type Recipe struct {
	Def  RcpDef
	Mats map[Material]int
}

func defParse(def string) (kind int, key string) {
	sep := strings.IndexRune(def, ':')
	if sep <= 0 {
		return -1, def
	}
	k, err := strconv.Atoi(def[:sep])
	if err != nil {
		return -1, def
	}
	return k, def[sep+1:]
}
