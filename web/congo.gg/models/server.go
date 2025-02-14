package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/google/uuid"
)

type Server struct {
	congo.Model
	UserID      string
	CheckoutID  string
	CheckoutURL string
	Name        string
	Size        string
	Status      serverStatus
	IpAddr      string
	Domain      string
	Error       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type serverStatus string

const (
	Paid      serverStatus = "paid"
	Launched  serverStatus = "launched"
	Prepared  serverStatus = "prepared"
	Assigned  serverStatus = "assigned"
	Ready     serverStatus = "ready"
	Destroyed serverStatus = "destroyed"
)

func (s *Server) StatusInt() int {
	switch s.Status {
	case Paid:
		return 1
	case Launched:
		return 2
	case Prepared:
		return 3
	case Assigned:
		return 4
	case Ready:
		return 5
	case Destroyed:
		return 6
	default:
		return 0
	}
}

func NewServer(db *congo.Database, userID, name, size string) (*Server, error) {
	id := strings.ToLower(strings.Replace(name, " ", "-", -1))
	_, err := GetServer(db, id)
	for err == nil {
		id = fmt.Sprintf("%s-%s", id, uuid.NewString()[:4])
		_, err = GetServer(db, id)
	}
	s := Server{Model: db.NewModel(id)}
	return &s, db.Query(`

		INSERT INTO servers (id, user_id, host_id, name, size)
		VALUES (?, ?, ?, ?, ?)
		RETURNING user_id, checkout_id, checkout_url, name, size, status, ip_addr, domain, error, created_at, updated_at
	
	`, s.ID, userID, "", name, size).Scan(&s.UserID, &s.CheckoutID, &s.CheckoutURL, &s.Name, &s.Size, &s.Status, &s.IpAddr, &s.Domain, &s.Error, &s.CreatedAt, &s.UpdatedAt)
}

func GetServer(db *congo.Database, id string) (*Server, error) {
	s := Server{Model: congo.Model{DB: db}}
	return &s, db.Query(`

		SELECT id, user_id, checkout_id, checkout_url, name, size, status, ip_addr, domain, error, created_at, updated_at
		FROM servers
		WHERE id = ?

	`, id).Scan(&s.ID, &s.UserID, &s.CheckoutID, &s.CheckoutURL, &s.Name, &s.Size, &s.Status, &s.IpAddr, &s.Domain, &s.Error, &s.CreatedAt, &s.UpdatedAt)
}

func ServersForUser(db *congo.Database, userID string) ([]*Server, error) {
	servers := make([]*Server, 0)
	return servers, db.Query(`

		SELECT id, user_id, checkout_id, checkout_url, name, size, status, ip_addr, domain, error, created_at, updated_at
		FROM servers
		WHERE user_id = ?

	`, userID).All(func(scan congo.Scanner) (err error) {
		s := Server{Model: congo.Model{DB: db}}
		servers = append(servers, &s)
		return scan(&s.ID, &s.UserID, &s.CheckoutID, &s.CheckoutURL, &s.Name, &s.Size, &s.Status, &s.IpAddr, &s.Domain, &s.Error, &s.CreatedAt, &s.UpdatedAt)
	})
}

func (s *Server) Save() error {
	return s.DB.Query(`

		UPDATE servers
		SET checkout_id = ?,
				checkout_url = ?,
				name = ?,
				size = ?,
				status = ?,
				ip_addr = ?,
				domain = ?,
				error = ?,
				updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		RETURNING updated_at

	`, s.CheckoutID, s.CheckoutURL, s.Name, s.Size, s.Status, s.IpAddr, s.Domain, s.Error, s.ID).Scan(&s.UpdatedAt)
}

func (s *Server) Delete() error {
	return s.DB.Query(`
	
		DELETE FROM servers
		WHERE id = ?
	
	`, s.ID).Exec()
}
