package models

import (
	"github.com/ccutch/congo/pkg/congo"
	"github.com/google/uuid"
)

type App struct {
	congo.Model
	Name   string
	Binary []byte
}

func NewApp(db *congo.Database, name string, binary []byte) (*App, error) {
	a := App{db.NewModel(uuid.NewString()), name, binary}
	return &a, db.Query(`
	
		INSERT INTO apps (id, name, binary)
		VALUES (?, ?, ?)
		RETURNING created_at, updated_at
	
	`, a.ID, a.Name, a.Binary).Scan(&a.CreatedAt, &a.UpdatedAt)
}

func AllApps(db *congo.Database) ([]*App, error) {
	apps := []*App{}
	return apps, db.Query(`

		SELECT id, name, binary, created_at, updated_at
		FROM apps
		ORDER BY created_at DESC
	
	`).All(func(scan congo.Scanner) error {
		a := App{Model: db.Model()}
		apps = append(apps, &a)
		return scan(&a.ID, &a.Name, &a.Binary, &a.CreatedAt, &a.UpdatedAt)
	})
}

func GetApp(db *congo.Database, id string) (*App, error) {
	a := App{Model: db.Model()}
	return &a, db.Query(`

		SELECT id, name, binary, created_at, updated_at
		FROM apps
		WHERE id = ?

	`, id).Scan(&a.ID, &a.Name, &a.Binary, &a.CreatedAt, &a.UpdatedAt)
}

func (app *App) Save() error {
	return app.DB.Query(`
	
		UPDATE apps
		SET name = ?,
				binary = ?,
				updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		RETURNING updated_at
	
	`, app.Name, app.Binary, app.ID).Scan(&app.UpdatedAt)
}

func (app *App) Delete() error {
	return app.DB.Query(`
	
		DELETE FROM apps
		WHERE id = ?
	
	`, app.ID).Exec()
}
