package main

import (
	"math"
)

// http://www.movable-type.co.uk/scripts/latlong.html
// latitude  φ N/S
// longitude λ E/W

func deg2rad(degree float64) (radian float64) {
	return degree * math.Pi / 180
}

func rad2deg(radian float64) (degree float64) {
	return degree * 180 / math.Pi
}

func dms2deg(deg, min, sec int) float64 {
	m := float64(min) / 60.0
	s := float64(min) / (60.0 * 60.0)
	return float64(deg) + m + s
}

func pow2(x float64) float64 { return x * x }

func sphereDistRad(r, lat1, lon1, lat2, lon2 float64) float64 {
	dp := lat2 - lat1 // φ2 - φ1
	dl := lon2 - lon1 // λ2 - λ1
	a := pow2(math.Sin(dp/2)) + math.Cos(lat1)*math.Cos(lat2)*pow2(math.Sin(dl/2))
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	d := r * c
	return d
}

func sphereDistDeg(r float64, lat1, lon1, lat2, lon2 float64) float64 {
	res := sphereDistRad(r, deg2rad(lat1), deg2rad(lon1), deg2rad(lat2), deg2rad(lon2))
	return res
}

func sphereBearingRad(lat1, lon1, lat2, lon2 float64) float64 {
	y := math.Sin(lon2-lon1) * math.Cos(lat2)
	x := math.Cos(lat1)*math.Sin(lat2) -
		math.Sin(lat1)*math.Cos(lat2)*math.Cos(lon2-lon1)
	b := math.Atan2(y, x)
	return b
}

func sphereBearingDeg(lat1, lon1, lat2, lon2 float64) float64 {
	res := sphereBearingRad(deg2rad(lat1), deg2rad(lon1), deg2rad(lat2), deg2rad(lon2))
	return res
}
