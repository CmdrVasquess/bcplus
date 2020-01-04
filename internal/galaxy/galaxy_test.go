package app

import (
	"fmt"
	"testing"
)

func TestGalaxy_putnget(t *testing.T) {
	g, err := OpenGalaxy("testgalaxy.bbolt")
	if err != nil {
		t.Fatal(err)
	}
	defer g.Close()
	sys := &System{EdId: 4711, Name: "Zeta Reticuli"}
	err = g.PutSystem(sys)
	if err != nil {
		t.Fatal(err)
	}
	sys, err = g.GetSystem(4711)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(sys)
}
