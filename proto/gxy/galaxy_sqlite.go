package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteDbSystem sql.DB

// TODO this probably needs some way of locking. No SELECT FOR UPDATE?!
func sqLiteSeqNext(tx *sql.Tx, ttag uint64) (res uint64, err error) {
	row := tx.QueryRow(`SELECT count FROM sequences WHERE ttag=$1`, ttag)
	if err = row.Scan(&res); err != nil {
		return 0, err
	}
	res++
	_, err = tx.Exec(`UPDATE sequences SET count=$1 WHERE ttag=$2`, res, ttag)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (dbsys *SQLiteDbSystem) Create(tx *sql.Tx, ttag uint64, e interface{}) (OId, error) {
	sys := e.(*System)
	seq, err := sqLiteSeqNext(tx, ttag)
	if err != nil {
		return 0, err
	}
	oid, err := MakeOId(ttag, seq)
	if err != nil {
		return 0, err
	}
	_, err = tx.Exec(`INSERT INTO system (id, name, cx, cy, cz)
	                  VALUES ($1, $2, $3, $4, $5)`,
		oid, sys.Name, sys.Coos[0], sys.Coos[1], sys.Coos[2])
	if err != nil {
		return 0, err
	}
	return oid, nil
}
