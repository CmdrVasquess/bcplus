package core

import (
	"os/user"
	"path/filepath"
)

var relJournalPath = []string{
	"",
	"Saved Games",
	"Frontier Developments",
	"Elite Dangerous",
}

func DefaultJournalDir() string {
	if usr, err := user.Current(); err != nil {
		return "."
	} else {
		relJournalPath[0] = usr.HomeDir
		res := filepath.Join(relJournalPath...)
		return res
	}
}

func DefaultDataDir() string {
	if usr, _ := user.Current(); usr == nil {
		return "."
	} else {
		return filepath.Join(usr.HomeDir, "bcplus")
	}
}
