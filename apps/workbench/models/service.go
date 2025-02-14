package models

import (
	"strings"

	"github.com/ccutch/congo/pkg/congo"
)

type Service struct {
	congo.Model
	Name   string
	Path   string
	Port   int
	Status string
	Error  string
}

func NewService(db *congo.Database, name, path string, port int) (*Service, error) {
	s := Service{
		Model: db.NewModel(strings.ToLower(strings.Replace(name, " ", "-", -1))),
		Name:  name,
		Path:  path,
		Port:  port,
	}
	return &s, db.Query(`

		INSERT INTO services (id, name, port, path)
		VALUES (?, ?, ?, ?)
		RETURNING created_at, updated_at

	`, s.ID, s.Name, s.Port, s.Path).Scan(&s.CreatedAt, &s.UpdatedAt)
}

func ListServices(db *congo.Database) ([]*Service, error) {
	services := []*Service{}
	return services, db.Query(`

		SELECT id, name, port, path, status, created_at, updated_at
		FROM services
		ORDER BY created_at DESC

	`).All(func(scan congo.Scanner) error {
		s := Service{Model: db.Model()}
		services = append(services, &s)
		return scan(&s.ID, &s.Name, &s.Port, &s.Path, &s.Status, &s.CreatedAt, &s.UpdatedAt)
	})
}

func GetService(db *congo.Database, id string) (*Service, error) {
	s := Service{Model: db.Model()}
	return &s, db.Query(`

		SELECT id, name, port, path, status, created_at, updated_at
		FROM services
		WHERE id = ?

	`, id).Scan(&s.ID, &s.Name, &s.Port, &s.Path, &s.Status, &s.CreatedAt, &s.UpdatedAt)
}

func (s *Service) Save() error {
	return s.DB.Query(`

		UPDATE services
		SET name = ?, path = ?, port = ?, status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		RETURNING updated_at

	`, s.Name, s.Path, s.Port, s.Status, s.ID).Scan(&s.UpdatedAt)
}

func (s *Service) Delete() error {
	return s.DB.Query(`

		DELETE FROM services
		WHERE id = ?

	`, s.ID).Exec()
}
