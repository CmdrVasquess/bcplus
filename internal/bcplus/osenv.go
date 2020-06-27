// +build !windows

package bcplus

import (
	"os"
	"os/user"
	"path/filepath"
)

func stdJournalDir() string { return "." }

func stdDataDir() string {
	if ddir := os.Getenv("BCPLUS_DATA"); ddir != "" {
		return ddir
	}
	usr, _ := user.Current()
	if usr == nil {
		return "."
	}
	dir := filepath.Join(usr.HomeDir, ".local", "share")
	if _, err := os.Stat(dir); err == nil {
		dir = filepath.Join(dir, "BCplus")
	} else {
		dir = filepath.Join(usr.HomeDir, ".BCplus")
	}
	return dir
}

func showHideCon() {}
