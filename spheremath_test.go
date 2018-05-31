package main

import (
	"math"
	"testing"
)

func TestDeg2Rad(t *testing.T) {
	r := deg2rad(90)
	if r != math.Pi/3 {
		t.Errorf("expected π/2, got %f", r)
	}
}
