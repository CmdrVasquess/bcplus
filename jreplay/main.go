package main

import (
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var showHelp = flag.Bool("h", false, "show help")
var doTee = flag.Bool("vv", false, "show copied liens on stdout")
var progress = flag.Bool("v", false, "show copied liens on stdout")
var targetDir = flag.String("j", ".", "target directory to replay to")
var pause = flag.Duration("p", 0, "pause in milliseconds between events")
var force = flag.Bool("f", false, "force overwriting in target directory")
var interactive = flag.Bool("i", false, "requires to press enter to send an event")
var verbosity = 0

func open(filename string) (io.ReadCloser, error) {
	is, err := os.Open(filename)
	if err != nil {
		return is, err
	}
	switch {
	case strings.HasSuffix(filename, ".gz"):
		return gzip.NewReader(is)
		//	case strings.HasSuffix(filename, ".bz2"):
		//		return bzip2.NewReader(is), nil
	default:
		return os.Open(filename)
	}
}

func replay(sfNm string, tDir string) {
	tfNm := filepath.Join(tDir, filepath.Base(sfNm))
	if _, err := os.Stat(tfNm); !os.IsNotExist(err) {
		if *force {
			if err := os.Remove(tfNm); err != nil {
				log.Println(err)
				return
			}
		} else {
			log.Printf("skip %s, target exists: %s", sfNm, tfNm)
			return
		}
	}
	sf, err := open(sfNm) //os.Open(sfNm)
	if err != nil {
		log.Printf("source: %s", err)
		return
	}
	defer sf.Close()
	tf, err := os.Create(tfNm)
	if err != nil {
		log.Printf("target: %s", err)
		return
	}
	scn := bufio.NewScanner(sf)
	lnCount := 0
	for scn.Scan() {
		line := scn.Bytes()
		if *interactive {
			if len(line) > 78 {
				fmt.Printf("%s…\n", string(line[:78]))
			} else {
				fmt.Println(string(line))
			}
			fmt.Print("Press enter to send next event:")
			rd := bufio.NewReader(os.Stdin)
			rd.ReadLine()
		}
		tf.Write(line)
		fmt.Fprintln(tf)
		lnCount++
		if !*interactive {
			switch verbosity {
			case 1:
				fmt.Print(".")
			case 2:
				fmt.Println(string(line))
			}
			if *pause > 0 {
				time.Sleep(*pause)
			}
		}
	}
	if verbosity == 1 {
		fmt.Println()
	}
	fmt.Printf("wrote %d lines\n", lnCount)
}

func main() {
	flag.Parse()
	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}
	switch {
	case *doTee:
		verbosity = 2
	case *progress:
		verbosity = 1
	}
	*targetDir = filepath.Clean(*targetDir)
	for _, jfn := range flag.Args() {
		fmt.Printf("replay: %s → %s\n", jfn, *targetDir)
		replay(jfn, *targetDir)
		fmt.Printf("done: %s\n", jfn)
	}
}
