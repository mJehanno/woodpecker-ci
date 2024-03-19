package db

import (
	"database/sql"
	_ "embed"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
	_ "modernc.org/sqlite"
)

var Conn *goqu.Database

//go:embed schema.sql
var schema string

func GetConnection() (*goqu.Database, error) {
	if Conn == nil {
		dialect := goqu.Dialect("sqlite3")

		db, err := sql.Open("sqlite", "pipelines.db")
		if err != nil {
			return nil, err
		}

		Conn = dialect.DB(db)
	}

	return Conn, nil
}

func CreateDB() error {
	con, err := GetConnection()
	if err != nil {
		return err
	}

	_, err = con.Exec(schema)
	if err != nil {
		return err
	}

	return nil
}
