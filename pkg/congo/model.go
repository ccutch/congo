package congo

import (
	"time"
)

type Model struct {
	DB        *Database
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (db *Database) Model() Model {
	return Model{DB: db}
}

func (db *Database) NewModel(id string) Model {
	return Model{DB: db, ID: id}
}
