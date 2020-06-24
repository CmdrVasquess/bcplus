package main

import (
	"database/sql"
	"sync"
)

type DbOps interface {
	Create(tx *sql.Tx, ttag uint64, e interface{}) (OId, error)
}

type DbApp struct {
	db    *sql.DB
	tx    *sql.Tx
	mutex sync.Mutex
}

type TxGuard bool

func (dba *DbApp) Atomic(autoCommit bool, do func(*sql.Tx) error) (err error) {
	tx, guard := dba.Tx()
	if autoCommit {
		defer guard.Commit(dba)
		return do(tx)
	} else {
		defer guard.Rollback(dba)
		if err = do(tx); err == nil {
			return guard.Commit(dba)
		}
		return err
	}
}

func (dba *DbApp) Tx() (*sql.Tx, TxGuard) {
	dba.mutex.Lock()
	defer dba.mutex.Unlock()
	if dba.tx == nil {
		var err error
		if dba.tx, err = dba.db.Begin(); err != nil {
			return nil, false
		}
		return dba.tx, true
	}
	return dba.tx, false
}

func (own TxGuard) Rollback(dba *DbApp) error {
	if own {
		dba.mutex.Lock()
		defer dba.mutex.Unlock()
		if dba.tx == nil {
			return nil
		}
		tx := dba.tx
		dba.tx = nil
		return tx.Rollback() // TODO log error
	}
	return nil
}

func (own TxGuard) Commit(dba *DbApp) error {
	if own {
		dba.mutex.Lock()
		defer dba.mutex.Unlock()
		if dba.tx == nil {
			return nil
		}
		tx := dba.tx
		dba.tx = nil
		return tx.Commit()
	}
	return nil
}
