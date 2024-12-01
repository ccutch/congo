package congo

import (
	"time"
)

type Model struct {
	*Database
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (db *Database) NewModel(id string) Model {
	return Model{
		Database: db,
		ID:       id,
	}
}
