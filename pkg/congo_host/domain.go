package congo_host

import (
	_ "embed"

	"github.com/ccutch/congo/pkg/congo"
)

type Domain struct {
	host *CongoHost
	congo.Model
	ServerID   string
	DomainName string
	Verified   bool
}

func (host *RemoteHost) Domain(name string) *Domain {
	return &Domain{
		host:       host.host,
		Model:      host.host.DB.NewModel(name),
		ServerID:   host.ID,
		DomainName: name,
	}
}

func (server *RemoteHost) Domains() ([]*Domain, error) {
	domains := []*Domain{}
	return domains, server.DB.Query(`

		SELECT id, server_id, domain_name, verified, created_at, updated_at
		FROM domains
		WHERE server_id = ?
		ORDER BY created_at

	`, server.ID).All(func(scan congo.Scanner) error {
		d := Domain{host: server.host, Model: server.host.DB.Model()}
		if err := scan(&d.ID, &d.ServerID, &d.DomainName, &d.Verified, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return err
		}
		domains = append(domains, &d)
		return nil
	})
}

func (host *CongoHost) GetDomain(domain string) (*Domain, error) {
	d := Domain{host: host, Model: host.DB.Model()}
	return &d, host.DB.Query(`

		SELECT id, server_id, domain_name, verified, created_at, updated_at
		FROM domains
		WHERE domain_name = ?

	`, domain).Scan(&d.ID, &d.ServerID, &d.DomainName, &d.Verified, &d.CreatedAt, &d.UpdatedAt)
}

func (domain *Domain) Save() error {
	return domain.DB.Query(`

		INSERT INTO domains (id, server_id, domain_name)
		VALUES (?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET
			verified = ?,
			updated_at = CURRENT_TIMESTAMP
		RETURNING updated_at

	`, domain.ID, domain.ServerID, domain.DomainName, domain.Verified).Scan(&domain.UpdatedAt)
}

func (domain *Domain) Delete() error {
	return domain.DB.Query(`

		DELETE FROM domains
		WHERE id = ?

	`, domain.ID).Exec()
}

func (domain *Domain) Server() (*RemoteHost, error) {
	return domain.host.GetServer(domain.ServerID)
}

func (domain *Domain) Verify(admin string) error {
	server, err := domain.Server()
	if err != nil {
		return err
	}

	domains, err := server.Domains()
	if err != nil {
		return err
	}

	otherDomains := []*Domain{domain}
	for _, d := range domains {
		if d.Verified {
			otherDomains = append(otherDomains, d)
		}
	}

	if err = server.Server.Verify(admin, otherDomains...); err != nil {
		return err
	}

	domain.Verified = true
	return domain.Save()
}
