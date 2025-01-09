package congo_host

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/pkg/errors"
)

type Domain struct {
	host *CongoHost
	congo.Model
	ServerID   string
	DomainName string
	Verified   bool
}

func (server *RemoteServer) NewDomain(name string) (*Domain, error) {
	domain := Domain{host: server.host, Model: server.host.DB.NewModel(name), ServerID: server.ID, DomainName: name}
	return &domain, server.host.DB.Query(`

		INSERT INTO domains (id, server_id, domain_name)
		VALUES (?, ?, ?)
		RETURNING created_at, updated_at

	`, domain.ID, server.ID, domain.DomainName).Scan(&domain.CreatedAt, &domain.UpdatedAt)
}

func (server *RemoteServer) Domains() ([]*Domain, error) {
	domains := []*Domain{}
	return domains, server.DB.Query(`

		SELECT id, server_id, domain_name, verified, created_at, updated_at
		FROM domains
		WHERE server_id = ?
		ORDER BY created_at DESC

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

		UPDATE domains
		SET verified = ?,
				updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		RETURNING updated_at

	`, domain.Verified, domain.ID).Scan(&domain.UpdatedAt)
}

func (domain *Domain) Delete() error {
	return domain.DB.Query(`

		DELETE FROM domains
		WHERE id = ?

	`, domain.ID).Exec()
}

func (domain *Domain) Server() (*RemoteServer, error) {
	return domain.host.GetServer(domain.ServerID)
}

//go:embed resources/server/generate-certs.sh
var generateCerts string

func (domain *Domain) Verify() error {
	server, err := domain.Server()
	if err != nil {
		return err
	}

	domains, err := server.Domains()
	if err != nil {
		return err
	}

	var existingDomains []string
	for _, d := range domains {
		if d.Verified {
			existingDomains = append(existingDomains, d.DomainName)
		}
	}

	id, other := domain.ID, strings.Join(existingDomains, " -d ")
	_, out, err := server.Run(nil, fmt.Sprintf(generateCerts, id, other))
	return errors.Wrap(err, out.String())
}
