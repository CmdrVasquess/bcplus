package galaxy

import (
	"math"

	"github.com/ungerik/go3d/float64/vec3"
)

type Vec3D = vec3.T

const (
	Xk = 0
	Yk = 1
	Zk = 2
)

var NaV3D Vec3D

func init() {
	V3dSet1(&NaV3D, math.NaN())
}

func V3dSet3(v *Vec3D, x, y, z float64) {
	v[Xk] = x
	v[Yk] = y
	v[Zk] = z
}

func V3dSet1(v *Vec3D, f float64) {
	v[Xk] = f
	v[Yk] = f
	v[Zk] = f
}

func V3dValid(v Vec3D) bool {
	return !math.IsNaN(v[Xk]) && !math.IsNaN(v[Yk]) && !math.IsNaN(v[Zk])
}
