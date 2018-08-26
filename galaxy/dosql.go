package galaxy

import (
	"bufio"
	"bytes"
	"database/sql"
	"os"
	"strings"
)

func runSqlFile(db *sql.DB, fnm string) error {
	sf, err := os.Open(fnm)
	if err != nil {
		return err
	}
	defer sf.Close()
	qbuf := bytes.NewBuffer(nil)
	scan := bufio.NewScanner(sf)
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if tx != nil {
			tx.Rollback()
		}
	}()
	for scan.Scan() {
		line := strings.Trim(scan.Text(), " \t")
		if len(line) == 0 && qbuf.Len() > 0 {
			_, err := db.Exec(qbuf.String())
			if err != nil {
				return err
			}
			qbuf.Reset()
		} else if !strings.HasPrefix(line, "--") {
			qbuf.WriteRune(' ')
			qbuf.WriteString(line)
		}
	}
	if qbuf.Len() > 0 {
		_, err := db.Exec(qbuf.String())
		if err != nil {
			return err
		}
	}
	tx.Commit()
	tx = nil
	return nil
}

func (rpo *Repo) RunSql(sqlFile string) error {
	return runSqlFile(rpo.db, sqlFile)
}
