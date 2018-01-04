// +build !windows

package main

import (
	"os/user"
	"path/filepath"
)

func defaultJournalDir() string {
	return "."
}

func defaultDataDir() string {
	if usr, _ := user.Current(); usr == nil {
		return "."
	} else {
		return filepath.Join(usr.HomeDir, ".bcplus")
	}
}
