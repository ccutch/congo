package models

import (
	"github.com/ccutch/congo/pkg/congo"
	"github.com/google/uuid"
)

type Host struct {
	congo.Model
	OwnerID string
	Name    string
	Size    string
	Region  string
	Error   string
	IpAddr  string
	Domain  string
}

func NewHost(db *congo.Database, ownerID, name, size, region string) (*Host, error) {
	h := Host{Model: db.Model()}
	return &h, db.Query(`

		INSERT INTO hosts (id, owner_id, name, size, region)
		VALUES (?, ?, ?, ?, ?)
		RETURNING id, owner_id, name, size, region, error, ip_addr, domain, created_at, updated_at
	
	`, uuid.NewString(), ownerID, name, size, region).
		Scan(&h.ID, &h.OwnerID, &h.Name, &h.Size, &h.Region, &h.Error, &h.IpAddr, &h.Domain, &h.CreatedAt, &h.UpdatedAt)
}

func GetHost(db *congo.Database, id string) (*Host, error) {
	h := Host{Model: db.Model()}
	return &h, db.Query(`
	
		SELECT id, owner_id, name, size, region, error, ip_addr, domain, created_at, updated_at
		FROM hosts
		WHERE id = ?
	
	`, id).
		Scan(&h.ID, &h.OwnerID, &h.Name, &h.Size, &h.Region, &h.Error, &h.IpAddr, &h.Domain, &h.CreatedAt, &h.UpdatedAt)
}

func SearchHosts(db *congo.Database, query string) (hosts []*Host, err error) {
	return hosts, db.Query(`
	
		SELECT id, owner_id, name, size, region, error, ip_addr, domain, created_at, updated_at
		FROM hosts
		WHERE title LIKE ?
	
	`, "%"+query+"%").All(func(scan congo.Scanner) (err error) {
		h := Host{Model: db.Model()}
		hosts = append(hosts, &h)
		return scan(&h.ID, &h.OwnerID, &h.Name, &h.Size, &h.Region, &h.Error, &h.IpAddr, &h.Domain, &h.CreatedAt, &h.UpdatedAt)
	})
}

func (h *Host) Save() error {
	return h.DB.Query(`
	
		UPDATE hosts
		SET error = ?,
				ip_addr = ?,
				domain = ?,
				updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		RETURNING updated_at

	`, h.Error, h.IpAddr, h.Domain, h.ID).Scan(&h.UpdatedAt)
}

func (h *Host) Delete() error {
	return h.DB.Query(`
	
		DELETE FROM hosts
		WHERE id = ?

	`, h.ID).Exec()
}
