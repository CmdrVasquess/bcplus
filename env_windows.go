package main

import (
	"fmt"
	"os/user"
	"path/filepath"
	"strings"
)

const winStdJDir = `C:\Users\%s\Saved Games\Frontier Developments\Elite Dangerous`

func defaultJournalDir() string {
	if usr, err := user.Current(); err != nil {
		return "."
	} else {
		unms := strings.Split(usr.Username, "\\")
		res := fmt.Sprintf(winStdJDir, unms[1])
		return res
	}
}

func defaultDataDir() string {
	if usr, _ := user.Current(); usr == nil {
		return "."
	} else {
		return filepath.Join(usr.HomeDir, "bcplus")
	}
}
