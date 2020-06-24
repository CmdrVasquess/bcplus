package main

import (
	"database/sql"
	"log"
)

const (
	TySystem uint64 = iota
)

func main() {
	sqlite, _ := sql.Open("sqlite3", "example.sqlite")
	defer sqlite.Close()
	db := DbApp{db: sqlite}
	sysRepo, _ := NewEtyRepo(&db, TySystem, (*SQLiteDbSystem)(nil))
	sol := &System{Name: "Sol", Coos: [3]float32{0, 0, 0}}
	err := db.Atomic(false, func(_ *sql.Tx) error {
		return sysRepo.Persist(sol)
	})
	if err != nil {
		log.Println(err)
	}
}
