package galaxy

import (
	"encoding/json"
	"os"
	"testing"
)

func TestGalaxy_putnget(t *testing.T) {
	g, err := OpenGalaxy("testgalaxy.bbolt")
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	sys := &System{
		SysDesc: SysDesc{Addr: 4711, Name: "Zeta Reticuli"},
		Center: &SysBody{
			Type: Star,
		},
	}
	err = g.PutSystem(sys)
	if err != nil {
		t.Fatal(err)
	}
	sys, err = g.FindSystem(4711)
	if err != nil {
		t.Fatal(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.Encode(sys)
}
