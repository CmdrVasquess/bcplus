package galaxy

import (
	"math"
)

const (
	Xk = 0
	Yk = 1
	Zk = 2
)

type Vec3D [3]float64

func (v *Vec3D) Valid() bool {
	nan := math.NaN()
	return v[Xk] != nan && v[Yk] != nan && v[Zk] != nan
}

func (v *Vec3D) Set(x, y, z float64) {
	v[Xk] = x
	v[Yk] = y
	v[Zk] = z
}

func (v *Vec3D) Set1(f float64) {
	v.Set(f, f, f)
}

func V3Dist2(lhs *Vec3D, rhs *Vec3D) float64 {
	d := rhs[Xk] - lhs[Xk]
	tmp := d * d
	d = rhs[Yk] - lhs[Yk]
	tmp += d * d
	d = rhs[Zk] - lhs[Zk]
	tmp += d * d
	return tmp
}

func V3Dist(lhs *Vec3D, rhs *Vec3D) float64 {
	tmp := V3Dist2(lhs, rhs)
	tmp = math.Sqrt(tmp)
	return tmp
}
