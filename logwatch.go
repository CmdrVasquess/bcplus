package main

/* To be portable we need to poll the logfile. On MS Win one only gets update
 * events, if the directory is "touched", i.e. a logfile that stays open and
 * regularly receives new content will not be notified until something happens
 * to its parent directory. E.g. pressing F5 in the file explorer helps –
 * but who want's to sit at the keyboard and press F5 from time to time??? */

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	str "strings"
	"time"

	"runtime"

	"github.com/fsnotify/fsnotify"
)

const sleepMax = 5000

// Unix: \n; Win: \r\n; Apple <= OS 9: \r
func splitLogLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if i := bytes.IndexAny(data, "\n\r"); i < 0 {
		return 0, nil, nil
	} else if len(data) == i+1 {
		return i + 1, data[0:i], nil
	} else if nc := data[i+1]; nc == '\n' || nc == '\r' {
		return i + 2, data[0:i], nil
	} else {
		return i + 1, data[0:i], nil
	}
}

func pollFile(watchFiles chan string, doPerLine func(line []byte)) {
	glog.Notice("file poller waiting for journals")
	var jrnlName string
	var jrnlFile *os.File
	var jrnlRdPos int64
	var sleep = 0
	defer func() {
		if jrnlFile != nil {
			jrnlFile.Close()
		}
	}()
	for {
		if len(jrnlName) == 0 {
			jrnlName = <-watchFiles
			if jrnlName == "" {
				glog.Info("exit logwatch file-poller")
				runtime.Goexit()
			}
			glog.Infof("start watching: %s", jrnlName)
			var err error
			if jrnlFile, err = os.Open(jrnlName); err != nil {
				glog.Errorf("cannot watch %s: %s", jrnlName, err)
				jrnlName = ""
			}
			jrnlRdPos = 0
			sleep = 0
		}
		jrnlStat, err := jrnlFile.Stat()
		if err != nil {
			glog.Errorf("cannot Stat() %s: %s", jrnlName, err)
			jrnlFile.Close()
			jrnlFile = nil
			jrnlName = ""
		} else {
			newRdPos := jrnlStat.Size()
			if newRdPos > jrnlRdPos {
				glog.Debugf("new bytes: %d [%d > %d]",
					newRdPos-jrnlRdPos,
					jrnlRdPos,
					newRdPos)
				jrnlScnr := bufio.NewScanner(jrnlFile)
				jrnlScnr.Split(splitLogLines)
				for jrnlScnr.Scan() {
					line := jrnlScnr.Bytes()
					doPerLine(line)
				}
				jrnlRdPos = newRdPos
				sleep = 0
			} else if len(watchFiles) == 0 {
				switch {
				case sleep == 0:
					sleep = 500
				case sleep < sleepMax:
					if sleep = 3 * sleep / 2; sleep > sleepMax {
						sleep = sleepMax
					}
				}
				glog.Debugf("nothing to do, sleep %d mSec…", sleep)
				time.Sleep(time.Duration(sleep) * time.Millisecond)
				glog.Debugf("…woke up again")
			} else {
				glog.Noticef("closing journal: %s", jrnlName)
				jrnlFile.Close()
				jrnlFile = nil
				jrnlName = ""
			}
		}
	}
}

func isJournalFile(name string) bool {
	return str.HasPrefix(name, "Journal.") &&
		str.HasSuffix(name, ".log")
}

func newestJournal(inDir string) (res string) {
	dir, err := os.Open(inDir)
	if err != nil {
		glog.Error("fail to scan journal-dir: ", err)
		return ""
	}
	defer dir.Close()
	var maxTime time.Time
	infos, err := dir.Readdir(1)
	for len(infos) > 0 && err == nil {
		info := infos[0]
		if isJournalFile(info.Name()) && (info.ModTime().After(maxTime) || len(res) == 0) {
			res = info.Name()
			maxTime = info.ModTime()
		}
		infos, err = dir.Readdir(1)
	}
	return filepath.Join(inDir, res)
}

func WatchJournal(done <-chan bool,
	pickupLatest bool,
	inDir string,
	doPerLine func([]byte)) {
	inDir = filepath.Clean(inDir)
	watch, err := fsnotify.NewWatcher()
	if err != nil {
		glog.Fatalf("cannot create fs-watcher: %s", err)
	}
	defer watch.Close()
	if err = watch.Add(inDir); err != nil {
		glog.Fatalf("cannot watch %s: %s", inDir, err)
	}
	watchList := make(chan string, 12) // do we really need backlog?
	go pollFile(watchList, doPerLine)
	if pickupLatest {
		if newest := newestJournal(inDir); len(newest) > 0 {
			glog.Infof("dispatching latest log: %s", newest)
			watchList <- newest
		}
	}
	glog.Infof("watching journals in: %s", inDir)
	for {
		select {
		case fse := <-watch.Events:
			if !isJournalFile(filepath.Base(fse.Name)) {
				glog.Debugf("ignore event %s on non-journal file: %s",
					fse.Op,
					fse.Name)
			} else if fse.Op&fsnotify.Create == fsnotify.Create {
				cleanName := filepath.Clean(fse.Name)
				glog.Debugf("enqueue new journal: %s", cleanName)
				watchList <- cleanName
			} else {
				glog.Debugf("ignore fs-event: %s @ %s", fse.Op, fse.Name)
			}
		case err = <-watch.Errors:
			glog.Errorf("fs-watch error: %q", err)
		case <-done:
			watchList <- ""
			glog.Info("exit journal watcher")
			runtime.Goexit()
		}
	}
}
