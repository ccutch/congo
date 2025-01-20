package models

import "github.com/ccutch/congo/pkg/congo"

func SetSetting(db *congo.Database, id string, val string) (err error) {
	return db.Query(`

		INSERT INTO settings (id, value)
		VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE SET
			value = excluded.value,
			updated_at = CURRENT_TIMESTAMP

	`, id, val).Exec()
}

func GetSetting(db *congo.Database, id string) (val string, err error) {
	return val, db.Query(`
	
		SELECT value
		FROM settings WHERE id = ?
	
	`, id).Scan(&val)
}
