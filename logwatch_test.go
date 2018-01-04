package main

import (
	"bufio"
	"fmt"
	"strings"
	"testing"
)

func TestSplitEmpty(t *testing.T) {
	rd := str.NewReader("")
	scn := bufio.NewScanner(rd)
	scn.Split(splitLogLines)
	scnCount := 0
	for scn.Scan() {
		scnCount++
	}
	if scnCount > 0 {
		t.Errorf("scanned too many lines: %d, expected 0", scnCount)
	}
}

func TestSplit1LineNoEndl(t *testing.T) {
	rd := str.NewReader("foo bar baz")
	scn := bufio.NewScanner(rd)
	scn.Split(splitLogLines)
	scnCount := 0
	for scn.Scan() {
		scnCount++
	}
	if scnCount > 0 {
		t.Errorf("scanned too many lines: %d, expected 0", scnCount)
	}
}

func ExampleSplit1LineNL() {
	rd := str.NewReader("foo bar baz\n")
	scn := bufio.NewScanner(rd)
	scn.Split(splitLogLines)
	for scn.Scan() {
		fmt.Printf("SCN:[%s]\n", scn.Text())
	}
	// Output:
	// SCN:[foo bar baz]
}

func ExampleSplit1LineCR() {
	rd := str.NewReader("foo bar baz\r")
	scn := bufio.NewScanner(rd)
	scn.Split(splitLogLines)
	for scn.Scan() {
		fmt.Printf("SCN:[%s]\n", scn.Text())
	}
	// Output:
	// SCN:[foo bar baz]
}

func ExampleSplit1LineCRNL() {
	rd := str.NewReader("foo bar baz\r\n")
	scn := bufio.NewScanner(rd)
	scn.Split(splitLogLines)
	for scn.Scan() {
		fmt.Printf("SCN:[%s]\n", scn.Text())
	}
	// Output:
	// SCN:[foo bar baz]
}

func ExampleSplit3LineNoCRNL() {
	rd := str.NewReader("foo\r\nbar\r\nbaz")
	scn := bufio.NewScanner(rd)
	scn.Split(splitLogLines)
	for scn.Scan() {
		fmt.Printf("SCN:[%s]\n", scn.Text())
	}
	// Output:
	// SCN:[foo]
	// SCN:[bar]
}

func ExampleSplit3LineCRNL() {
	rd := str.NewReader("foo\r\nbar\r\nbaz\n")
	scn := bufio.NewScanner(rd)
	scn.Split(splitLogLines)
	for scn.Scan() {
		fmt.Printf("SCN:[%s]\n", scn.Text())
	}
	// Output:
	// SCN:[foo]
	// SCN:[bar]
	// SCN:[baz]
}
