// +build !windows

package core

import (
	"os/user"
	"path/filepath"
)

func DefaultJournalDir() string {
	return "."
}

func DefaultDataDir() string {
	if usr, _ := user.Current(); usr == nil {
		return "."
	} else {
		return filepath.Join(usr.HomeDir, ".bcplus")
	}
}
