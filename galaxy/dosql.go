package galaxy

import (
	"bufio"
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func runSqlFile(db *sql.DB, updFrom int, fnm string) error {
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
	since := 0
	const sincePrefix = "-- since "
	for scan.Scan() {
		line := strings.Trim(scan.Text(), " \t")
		switch {
		case len(line) == 0 && qbuf.Len() > 0:
			if since > updFrom {
				_, err := db.Exec(qbuf.String())
				if err != nil {
					return err
				}
			}
			qbuf.Reset()
		case strings.HasPrefix(line, sincePrefix):
			since, err = strconv.Atoi(strings.TrimSpace(line[len(sincePrefix):]))
			if err != nil {
				return fmt.Errorf("cannot pase version in since-line: '%s'", line)
			}
		case !strings.HasPrefix(line, "--"):
			qbuf.WriteRune(' ')
			qbuf.WriteString(line)
		}
	}
	if qbuf.Len() > 0 && since > updFrom {
		_, err := db.Exec(qbuf.String())
		if err != nil {
			return err
		}
	}
	tx.Commit()
	tx = nil
	return nil
}

func (rpo *Repo) RunSql(updFrom int, sqlFile string) error {
	return runSqlFile(rpo.db, updFrom, sqlFile)
}
