package models

import (
	"github.com/ccutch/congo/pkg/congo"
	"github.com/google/uuid"
)

type Host struct {
	congo.Model
	OwnerID    string
	ServerID   string
	PaymentID  string
	Name       string
	DomainName string
	Status     string
	Error      string
}

func NewHost(db *congo.Database, ownerID, paymentID, name, domainName string) (*Host, error) {
	h := Host{
		Model:      db.NewModel(uuid.NewString()),
		OwnerID:    ownerID,
		PaymentID:  paymentID,
		Name:       name,
		DomainName: domainName,
	}
	return &h, db.Query(`

		INSERT INTO hosts (id, owner_id, server_id, payment_id, name, domain_name)
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING status, created_at, updated_at

	`, h.ID, h.OwnerID, h.PaymentID, h.PaymentID, h.Name, h.DomainName).Scan(&h.Status, &h.CreatedAt, &h.UpdatedAt)
}

func GetHost(db *congo.Database, id string) (*Host, error) {
	h := Host{Model: db.NewModel(id)}
	return &h, db.Query(`
	
		SELECT owner_id, server_id, payment_id, name, domain_name, status, error, created_at, updated_at
		FROM hosts
		WHERE id = ?

	`, id).Scan(&h.OwnerID, &h.ServerID, &h.PaymentID, &h.Name, &h.DomainName, &h.Status, &h.Error, &h.CreatedAt, &h.UpdatedAt)
}

func HostsForOwner(db *congo.Database, ownerID string) (hosts []*Host, err error) {
	return hosts, db.Query(`
	
		SELECT id, owner_id, server_id, payment_id, name, domain_name, status, error, created_at, updated_at
		FROM hosts
		WHERE owner_id = ?
	
	`, ownerID).All(func(scan congo.Scanner) error {
		h := Host{Model: db.Model()}
		hosts = append(hosts, &h)
		return scan(&h.ID, &h.OwnerID, &h.ServerID, &h.PaymentID, &h.Name, &h.DomainName, &h.Status, &h.Error, &h.CreatedAt, &h.UpdatedAt)
	})
}

func (h *Host) Save() error {
	return h.DB.Query(`

		UPDATE hosts
		SET owner_id = ?,
				server_id = ?,
				payment_id = ?,
				name = ?,
				domain_name = ?,
				status = ?,
				error = ?,
				updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		RETURNING updated_at

	`, h.OwnerID, h.ServerID, h.PaymentID, h.Name, h.DomainName, h.Status, h.Error, h.ID).Scan(&h.UpdatedAt)
}

func (h *Host) Delete() error {
	return h.DB.Query(`

		DELETE FROM hosts
		WHERE id = ?

	`, h.ID).Exec()
}
