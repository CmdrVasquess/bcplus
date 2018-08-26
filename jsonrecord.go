package main

import (
	"github.com/CmdrVasquess/BCplus/galaxy"
)

func jvuV3d(v *galaxy.Vec3D, jv interface{}) {
	tmp := jv.([]interface{})
	v[galaxy.Xk] = tmp[0].(float64)
	v[galaxy.Yk] = tmp[1].(float64)
	v[galaxy.Zk] = tmp[2].(float64)
}

func jvgV3d(jv interface{}) (res galaxy.Vec3D) {
	jvuV3d(&res, jv)
	return res
}
