package congo

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"

	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	*sql.DB
	migrations *migrate.Migrate
	Root       string
}

func SetupDatabase(root, name string, migrations fs.FS) *Database {
	db := Database{Root: root}
	err := os.MkdirAll(root, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}
	dbFilePath := filepath.Join(root, name)
	if db.DB, err = sql.Open("sqlite3", fmt.Sprintf("file:%s", dbFilePath)); err != nil {
		log.Fatalf("Failed to connect to datatabase: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	fs, err := iofs.New(migrations, "migrations")
	if err != nil {
		log.Fatalf("failed to find migrations: %s", err)
	}
	dest := fmt.Sprintf("sqlite3://%s/"+name, root)
	if db.migrations, err = migrate.NewWithSourceInstance("iofs", fs, dest); err != nil {
		log.Fatalf("failed to parse migrations: %s", err)
	}
	return &db
}

func (db *Database) MigrateDown() error {
	err := db.migrations.Down()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func (db *Database) MigrateUp() error {
	err := db.migrations.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func (db *Database) Query(q string, args ...any) *query {
	return &query{db, q, args}
}

type query struct {
	*Database
	text string
	args []any
}

type Scanner func(...any) error
type Reader func(Scanner) error

func (query *query) Exec() error {
	_, err := query.DB.Exec(query.text, query.args...)
	return err
}

func (query *query) Scan(args ...any) error {
	row := query.DB.QueryRow(query.text, query.args...)
	if err := row.Err(); err != nil {
		return err
	}
	return row.Scan(args...)
}

func (query *query) One(fn Reader) error {
	row := query.DB.QueryRow(query.text, query.args...)
	if err := row.Err(); err != nil {
		return err
	}
	return fn(row.Scan)
}

func (query *query) All(fn Reader) error {
	rows, err := query.DB.Query(query.text, query.args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		if err := fn(rows.Scan); err != nil {
			return err
		}
	}
	return nil
}

func (query *query) Page(limit int, fn Reader) (more bool, err error) {
	rows, err := query.DB.Query(query.text, query.args...)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	var count int
	for rows.Next() {
		if count++; count == limit+1 {
			return true, nil
		}
		if err := fn(rows.Scan); err != nil {
			return false, err
		}
	}
	return false, nil
}
