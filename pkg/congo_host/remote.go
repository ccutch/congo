package congo_host

import (
	_ "embed"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/pkg/errors"
)

type RemoteHost struct {
	Server
	congo.Model
	host   *CongoHost
	Name   string
	Size   string
	Region string

	Stdin  io.Reader
	Stdout io.Writer
}

func (host *CongoHost) NewServer(name, size, region string) (*RemoteHost, error) {
	if host.api == nil {
		return nil, errors.New("no platform provided")
	}
	s := RemoteHost{
		Server: host.api.Server(name),
		Model:  host.DB.NewModel(name),
		host:   host,
		Name:   name,
		Size:   size,
		Region: region,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
	}
	return &s, host.DB.Query(`
	
	  INSERT INTO servers (id, name, size, location)
		VALUES (?, ?, ?, ?)
		RETURNING created_at, updated_at
	
	`, s.ID, s.Name, s.Size, s.Region).Scan(&s.CreatedAt, &s.UpdatedAt)
}

func (host *CongoHost) GetServer(id string) (*RemoteHost, error) {
	log.Println("api", host.api)
	s := RemoteHost{
		Model:  host.DB.Model(),
		Server: host.api.Server(id),
		host:   host,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
	}
	return &s, host.DB.Query(`

		SELECT id, name, size, location, created_at, updated_at
		FROM servers
		WHERE id = ?

	`, id).Scan(&s.ID, &s.Name, &s.Size, &s.Region, &s.CreatedAt, &s.UpdatedAt)
}

func (host *CongoHost) ListServers() (servers []*RemoteHost, err error) {
	return servers, host.DB.Query(`

		SELECT id, name, size, location, created_at, updated_at
		FROM servers
		ORDER BY created_at DESC

	`).All(func(scan congo.Scanner) error {
		s := RemoteHost{Model: host.DB.Model(), host: host, Stdin: os.Stdin, Stdout: os.Stdout}
		err = scan(&s.ID, &s.Name, &s.Size, &s.Region, &s.CreatedAt, &s.UpdatedAt)
		servers = append(servers, &s)
		s.Server = host.api.Server(s.ID)
		return err
	})
}

func (h *RemoteHost) Save() error {
	return h.host.DB.Query(`

		UPDATE servers
		SET name = ?,
				size = ?,
				location = ?,
				updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		RETURNING updated_at

	`, h.Name, h.Size, h.Region, h.ID).Scan(&h.UpdatedAt)
}

func (h *RemoteHost) Delete(purge, force bool) error {
	if err := h.Server.Delete(purge, force); err != nil {
		return errors.Wrap(err, "failed to delete server")
	}
	return h.host.DB.Query(`
	
		DELETE FROM servers WHERE id = ?
		
	`, h.ID).Exec()
}

func (h *RemoteHost) SetStdin(stdin io.Reader) {
	h.Stdin = stdin
}

func (h *RemoteHost) SetStdout(stdout io.Writer) {
	h.Stdout = stdout
}

func (h *RemoteHost) Run(args ...string) error {
	return h.Server.Run(h.Stdin, h.Stdout, args...)
}

//go:embed resources/server/prepare-server.sh
var prepareServer string

func (h *RemoteHost) Prepare() error {
	return h.Run(fmt.Sprintf(prepareServer, h.Name, h.Size, h.Region))
}

//go:embed resources/server/start-server.sh
var startServer string

func (h *RemoteHost) Restart() error {
	return h.Run(fmt.Sprintf(startServer, 8080))
}

func (h *RemoteHost) Deploy(source string) error {
	if _, out, err := h.Copy(source, "/root/congo"); err != nil {
		return errors.Wrap(err, "failed to copy source to server: "+out.String())
	}
	return h.Restart()
}
