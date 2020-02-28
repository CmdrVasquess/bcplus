package app

import (
	"math"
	"testing"
)

func TestTrack(t *testing.T) {
	type coo struct{ lat, lon float64 }
	berlin := coo{lat: 52.517, lon: 13.4}
	tokyo := coo{lat: 35.7, lon: 139.767}
	track := Track(6370,
		DegToRad(berlin.lat), DegToRad(berlin.lon),
		DegToRad(tokyo.lat), DegToRad(tokyo.lon),
	)
	if math.Abs(track-8918) > .5 {
		t.Errorf("want track 8918, got: %.2f", track)
	}
}

// func TestTrackAngle(t *testing.T) {
// 	type coo struct{ lat, lon float64 }
// 	berlin := coo{lat: 52.517, lon: 13.4}
// 	tokyo := coo{lat: 35.7, lon: 139.767}
// 	// berlin := coo{lat: 0, lon: 0}
// 	// tokyo := coo{lat: -10, lon: -10}
// 	a := RadToDeg(TrackAngle(
// 		DegToRad(berlin.lat), DegToRad(berlin.lon),
// 		DegToRad(tokyo.lat), DegToRad(tokyo.lon),
// 	))
// 	//if math.Abs(80.212-a) >= 0 {
// 	t.Errorf("want track angle 80.212, got: %.2f", a)
// 	//}
// }

func TestBearing(t *testing.T) {
	type coo struct{ lat, lon float64 }
	a := RadToDeg(Bearing(0, 0, DegToRad(10), DegToRad(0)))
	if a != 0 {
		t.Errorf("want track angle 0, got: %.2f", a)
	}
	a = RadToDeg(Bearing(0, 0, DegToRad(0), DegToRad(10)))
	if a != 90 {
		t.Errorf("want track angle 90, got: %.2f", a)
	}
	a = RadToDeg(Bearing(0, 0, DegToRad(-10), DegToRad(0)))
	if a != 180 {
		t.Errorf("want track angle 180, got: %.2f", a)
	}
	a = RadToDeg(Bearing(0, 0, DegToRad(0), DegToRad(-10)))
	if a != 270 {
		t.Errorf("want track angle 270, got: %.2f", a)
	}
}
