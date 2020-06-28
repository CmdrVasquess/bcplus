package bcplus

import (
	"os"
	"os/user"
	"path/filepath"
	//"github.com/gonutz/w32"
)

var relJournalPath = []string{
	"",
	"Saved Games",
	"Frontier Developments",
	"Elite Dangerous",
}

func stdJournalDir() string {
	if usr, err := user.Current(); err != nil {
		return "."
	} else {
		relJournalPath[0] = usr.HomeDir
		res := filepath.Join(relJournalPath...)
		return res
	}
}

func stdDataDir() string {
	if ddir := os.Getenv("BCPLUS_DATA"); ddir != "" {
		return ddir
	}
	usr, _ := user.Current()
	if usr == nil {
		return "."
	}
	dir := filepath.Join(usr.HomeDir, "AppData", "Roaming")
	if _, err := os.Stat(dir); err == nil {
		dir = filepath.Join(dir, "BCplus")
	} else {
		dir = filepath.Join(usr.HomeDir, "BCplus")
	}
	return dir
}

// var flagShowCon bool

// func init() {
// 	flag.BoolVar(&flagShowCon, "show-con", false, "show console window")
// }

// func showHideCon() {
// 	if !flagShowCon {
// 		console := w32.GetConsoleWindow()
// 		if console != 0 {
// 			_, consoleProcID := w32.GetWindowThreadProcessId(console)
// 			if w32.GetCurrentProcessId() == consoleProcID {
// 				w32.ShowWindowAsync(console, w32.SW_HIDE)
// 			}
// 		}
// 	}
// }
