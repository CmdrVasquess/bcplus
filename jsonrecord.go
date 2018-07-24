package main

import (
	"fmt"
	"time"

	"github.com/CmdrVasquess/BCplus/galaxy"
)

func jrgBool(jr map[string]interface{}, name string, fallback bool) bool {
	jv, ok := jr[name]
	if ok {
		return jv.(bool)
	}
	return fallback
}

func jrgInt(jr map[string]interface{}, name string, fallback int) int {
	jv, ok := jr[name]
	if ok {
		return int(jv.(float64))
	}
	return fallback
}

func jruInt(dst *int, jr map[string]interface{}, name string) {
	jv, ok := jr[name]
	if ok {
		*dst = int(jv.(float64))
	}
}

func jruInt64(dst *int64, jr map[string]interface{}, name string) {
	jv, ok := jr[name]
	if ok {
		*dst = int64(jv.(float64))
	}
}

func jruF32(dst *float32, jr map[string]interface{}, name string) {
	jv, ok := jr[name]
	if ok {
		*dst = float32(jv.(float64))
	}
}
func jrgStr(jr map[string]interface{}, name string, fallback string) string {
	jv, ok := jr[name]
	if ok {
		return jv.(string)
	}
	return fallback
}

func jrmgStr(jr map[string]interface{}, name string) string {
	jv, ok := jr[name]
	if ok {
		return jv.(string)
	}
	panic(fmt.Errorf("no string attribute '%s' in event", name))
}

func jruStr(dst *string, jr map[string]interface{}, name string) {
	jv, ok := jr[name]
	if ok {
		*dst = jv.(string)
	}
}

func jrgTs(jr map[string]interface{}, name string, fallback time.Time) (time.Time, error) {
	jv, ok := jr[name]
	if ok {
		res, err := time.Parse(time.RFC3339, jv.(string))
		if err != nil {
			return fallback, err
		}
		return res, nil
	}
	return fallback, nil
}

func jrmgTs(jr map[string]interface{}, name string) time.Time {
	res, err := jrgTs(jr, name, time.Time{})
	if err != nil {
		panic(err)
	}
	return res
}

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
