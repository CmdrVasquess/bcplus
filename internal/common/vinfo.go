package common

import (
	"fmt"
	"runtime"
)

var (
	VersionShort = fmt.Sprintf("%d.%d.%d", BCpMajor, BCpMinor, BCpPatch)
	VersionLong  = fmt.Sprintf("v%s-%s+%d", VersionShort, BCpQuality, BCpBuildNo)
	VersionAll   = fmt.Sprintf("%s (%s) / %s", VersionLong, BCpDate, runtime.Version())
)
