// +build !windows

package main

import (
	"fmt"
	"io"
)

const (
	BCpMajor  uint16 = 0
	BCpMinor  uint16 = 1
	BCpBugfix uint16 = 3
	BCpDate   string = "dev"
)

func BCpDescribe(wr io.Writer) {
	fmt.Fprintf(wr, "fractal[qb]: BC+ v%d.%d.%d (%s)",
		BCpMajor,
		BCpMinor,
		BCpBugfix,
		BCpDate)
}
