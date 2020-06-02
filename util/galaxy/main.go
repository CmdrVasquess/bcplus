package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"time"

	"git.fractalqb.de/fractalqb/ggja"

	_ "github.com/mattn/go-sqlite3"
)

var (
	flagDb   string
	flagHist bool
	stats    struct {
		fline     int
		lines     int
		jumps     int
		sysnoaddr int
		sysfnaddr int
	}
	cmdrId     int
	startafter time.Time
	dnkCmdrs   = make(map[string]int)
	dnkSyss    = make(map[string]int64)
)

func init() {
	flag.StringVar(&flagDb, "db", "galaxy.db", "sqlite3 DB file")
	flag.BoolVar(&flagHist, "historic", false, "also process old events")
}

var (
	eventSep = []byte(`"event":"`)
	eSepLen  = len(eventSep)
)

func jEvent(line []byte) string {
	idx := bytes.Index(line, eventSep)
	if idx < 0 {
		return ""
	}
	line = line[idx+eSepLen:]
	if idx = bytes.IndexByte(line, '"'); idx < 0 {
		log.Fatal("unterminated event name")
	}
	return string(line[:idx])
}

var evtHdlrs = map[string]func(time.Time, ggja.Obj, *sql.Tx){
	"LoadGame": hdlLoadGame,
	//"Location":  hdlLocation,
	"FSDJump": hdlFSDJump,
}

func reidCmdr(from, to int, tx *sql.Tx) {
	_, err := tx.Exec(`UPDATE cmdr SET id=? WHERE id=?`, to, from)
	if err != nil {
		log.Fatal(err)
	}
	_, err = tx.Exec(`UPDATE visit SET cmdr=? WHERE cmdr=?`, to, from)
	if err != nil {
		log.Fatal(err)
	}
}

func hdlLoadGame(t time.Time, evt ggja.Obj, tx *sql.Tx) {
	fid := evt.Str("FID", "")
	name := evt.MStr("Commander")
	if fid == "" {
		err := tx.QueryRow(`SELECT id FROM cmdr WHERE name=?`, name).Scan(&cmdrId)
		if err == sql.ErrNoRows {
			if cmdrId = dnkCmdrs[name]; cmdrId >= 0 {
				cmdrId = -(len(dnkCmdrs) + 1)
				_, err = tx.Exec(`INSERT INTO cmdr (id, name) VALUES (?,?)`,
					cmdrId, name)
				if err != nil {
					log.Panic(err)
				}
				dnkCmdrs[name] = cmdrId
			}
		} else if err != nil {
			log.Fatal(err)
		} else {
			dnkCmdrs[name] = cmdrId
		}
	} else {
		var err error
		cmdrId, err = strconv.Atoi(fid[1:])
		if err != nil {
			log.Panic(err)
		}
		if ukid := dnkCmdrs[name]; ukid < 0 {
			reidCmdr(ukid, cmdrId, tx)
			dnkCmdrs[name] = cmdrId
		} else if ukid == 0 {
			_, err = tx.Exec(`INSERT INTO cmdr (id, name) VALUES (?,?)`,
				cmdrId, name)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func hdlLocation(t time.Time, evt ggja.Obj, tx *sql.Tx) {
	addr := evt.Uint64("SystemAddress", 0)
	name := evt.Str("StarSystem", "")
	var err error
	if addr == 0 {
		return
	}
	_, err = tx.Exec(`INSERT INTO sys (addr, name) VALUES (?, ?)
	                  ON CONFLICT (addr) DO UPDATE SET name=?`,
		addr, name, name)
	if err != nil {
		log.Panic(err)
	}
}

func reidSys(from, to int64, name string, tx *sql.Tx) {
	_, err := tx.Exec(`INSERT INTO sys (addr, name, x,y,z)
	                   SELECT ?, '-rename-', x,y,z FROM sys
			           WHERE addr=?`,
		to, from)
	if err != nil {
		log.Panic(err)
	}
	_, err = tx.Exec(`UPDATE sysloc SET sys=? where sys=?`, to, from)
	if err != nil {
		log.Panic(err)
	}
	_, err = tx.Exec(`UPDATE visit SET sys=? where sys=?`, to, from)
	if err != nil {
		log.Panic(err)
	}
	_, err = tx.Exec(`DELETE FROM sys WHERE addr=?`, from)
	if err != nil {
		log.Panic(err)
	}
	_, err = tx.Exec(`UPDATE sys SET name=? WHERE addr=?`, name, to)
	if err != nil {
		log.Panic(err)
	}
}

func hdlFSDJump(t time.Time, evt ggja.Obj, tx *sql.Tx) {
	addr := evt.Int64("SystemAddress", 0)
	name := evt.MStr("StarSystem")
	coos := evt.MArr("StarPos")
	defer func() {
		if p := recover(); p != nil {
			log.Printf("panic FSDJump: %s (%d): %s", name, addr, p)
			panic(p)
		}
	}()
	if addr == 0 {
		stats.sysnoaddr++
		err := tx.QueryRow(`SELECT addr FROM sys WHERE name=?`, name).Scan(&addr)
		if err == sql.ErrNoRows {
			if addr = dnkSyss[name]; addr >= 0 {
				addr = int64(-(len(dnkSyss) + 1))
			}
		} else if err != nil {
			log.Fatal(err)
		} else if addr > 0 {
			delete(dnkSyss, name)
		}
	} else if ukAddr := dnkSyss[name]; ukAddr < 0 {
		stats.sysfnaddr++
		reidSys(ukAddr, addr, name, tx)
		delete(dnkSyss, name)
	}
	var err error
	_, err = tx.Exec(`INSERT INTO sys (addr, name, x,y,z)
	                  VALUES (?, ?, ?,?,?)
					  ON CONFLICT (addr) DO UPDATE
					  SET name=?, x=?,y=?,z=?`,
		addr, name, coos.MF64(0), coos.MF64(1), coos.MF64(2),
		name, coos.MF64(0), coos.MF64(1), coos.MF64(2))
	if err != nil {
		log.Panic(err)
	}
	if addr < 0 {
		dnkSyss[name] = addr
	}
	if cmdrId != 0 {
		_, err = tx.Exec(`INSERT INTO visit (cmdr, sys, t)
		                  VALUES (?, ?, ?)`,
			cmdrId, addr, t)
		if err != nil {
			log.Fatal(err)
		}
	}
	stats.jumps++
}

func importFrom(rd io.Reader, tx *sql.Tx) {
	stats.fline = 0
	scn := bufio.NewScanner(rd)
	for scn.Scan() {
		stats.fline++
		line := scn.Bytes()
		if h := evtHdlrs[jEvent(line)]; h != nil {
			evt := ggja.Obj{
				Bare:    make(ggja.GenObj),
				OnError: func(err error) { log.Panic(err) },
			}
			if err := json.Unmarshal(line, &evt.Bare); err != nil {
				log.Panic(err)
			}
			t := evt.MTime("timestamp")
			if t.After(startafter) {
				h(t, evt, tx)
			}
		}
		stats.lines++
	}
}

func importFile(name string, tx *sql.Tx) (err error) {
	rd, err := os.Open(name)
	if err != nil {
		log.Panic(err)
	}
	defer func() {
		rd.Close()
		if p := recover(); p != nil {
			err = fmt.Errorf("%s:%d:%v", name, stats.fline, p)
		}
	}()
	importFrom(rd, tx)
	return nil
}

func loadCmdrs(db *sql.DB) {
	rows, err := db.Query(`SELECT id, name FROM cmdr`)
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var id int
		var name string
		if err = rows.Scan(&id, &name); err != nil {
			log.Fatal(err)
		}
		dnkCmdrs[name] = id
	}
}

func loadDnkSyss(db *sql.DB) {
	rows, err := db.Query(`SELECT addr, name FROM sys WHERE addr < 0`)
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var addr int64
		var name string
		if err = rows.Scan(&addr, &name); err != nil {
			log.Fatal(err)
		}
		dnkSyss[name] = addr
	}
}

func latestVisit(db *sql.DB) time.Time {
	var tmp string
	err := db.QueryRow(`SELECT max(t) FROM visit`).Scan(&tmp)
	if err != nil {
		log.Fatal(err)
	}
	res, err := time.Parse("2006-01-02 15:04:05Z07:00", tmp)
	if err != nil {
		log.Fatal(err)
	}
	return res
}

func main() {
	flag.Parse()
	db, err := sql.Open("sqlite3", flagDb)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	loadCmdrs(db)
	log.Printf("loaded %d commanders\n", len(dnkCmdrs))
	loadDnkSyss(db)
	log.Printf("loaded %d unkonwn systems", len(dnkSyss))
	startafter = latestVisit(db)
	files := flag.Args()
	sort.Strings(files)
	for _, file := range files {
		log.Println("import", file)
		tx, err := db.Begin()
		if err != nil {
			log.Fatal(err)
		}
		if err = importFile(file, tx); err != nil {
			tx.Rollback()
			log.Fatal(err)
		} else {
			tx.Commit()
		}
	}
	log.Printf("stats: %#v\n", stats)
}
