package app

import (
	"math"
)

// https://de.wikipedia.org/wiki/Orthodrome

const deg2rad = math.Pi / 180
const rad2deg = 180 / math.Pi

func DegToRad(d float64) float64 { return d * deg2rad }
func RadToDeg(r float64) float64 { return r * rad2deg }

func Track1(p1, l1, p2, l2 float64) float64 {
	sp1, cp1 := math.Sincos(p1)
	sp2, cp2 := math.Sincos(p2)
	cl21 := math.Cos(l2 - l1)
	return math.Acos(sp1*sp2 + cp1*cp2*cl21)
}

func Track(r, p1, l1, p2, l2 float64) float64 {
	return r * Track1(p1, l1, p2, l2)
}

// func TrackAngle(p1, l1, p2, l2 float64) float64 {
// 	sp1, cp1 := math.Sincos(p1)
// 	sp2, cp2 := math.Sincos(p2)
// 	z := math.Acos(sp1*sp2 + cp1*cp2*math.Cos(l2-l1))
// 	q := (sp2 - sp1*math.Cos(z)) / (cp1 * math.Sin(z))
// 	if q <= -1 {
// 		return math.Pi
// 	} else if q >= 1 {
// 		return 0
// 	}
// 	return math.Acos(q)
// }

func Bearing(p1, l1, p2, l2 float64) (rad float64) {
	y := math.Sin(l1-l2) * math.Cos(p1)
	x := math.Cos(p2)*math.Sin(p1) -
		math.Sin(p2)*math.Cos(p1)*math.Cos(l1-l2)
	rad = math.Atan2(y, x) + math.Pi
	if rad >= 2*math.Pi {
		return rad - 2*math.Pi
	}
	return rad
}

func BearingDeg(p1, l1, p2, l2 float64) float64 {
	return RadToDeg(Bearing(
		DegToRad(p1), DegToRad(l1),
		DegToRad(p2), DegToRad(l2),
	))
}
