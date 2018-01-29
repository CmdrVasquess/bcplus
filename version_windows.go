package main

import (
	"bytes"
	"fmt"
	"io"
)

const (
	BCpMajor  uint16 = 0
	BCpMinor  uint16 = 4
	BCpBugfix uint16 = 3
	BCpDate   string = "Mo 29. Jan 21:02:58 CET 2018"
)

func BCpDescribe(wr io.Writer) {
	fmt.Fprintf(wr, "fractal[qb]: BC+ v%d.%d.%d (%s)",
		BCpMajor,
		BCpMinor,
		BCpBugfix,
		BCpDate)
}

func BCpDescStr() string {
	buf := bytes.NewBuffer(nil)
	BCpDescribe(buf)
	return buf.String()
}
