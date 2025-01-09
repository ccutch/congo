package congo_host

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/pkg/errors"
)

type Target interface {
	Run(io.Reader, ...string) (stdout, stderr bytes.Buffer, _ error)
}

type LocalServer struct {
	host *CongoHost
}

func (host *CongoHost) Local() *LocalServer {
	return &LocalServer{host: host}
}

func (s *LocalServer) Run(stdin io.Reader, args ...string) (stdout bytes.Buffer, stderr bytes.Buffer, err error) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = stdin
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	return stdout, stderr, cmd.Run()
}

type RemoteServer struct {
	host *CongoHost
	Server
	congo.Model
	Name     string
	Size     string
	Location string
	IP       string

	Stdin  io.Reader
	Stdout io.Writer
}

func (host *CongoHost) NewServer(name, size, location string) (*RemoteServer, error) {
	s := RemoteServer{
		host:     host,
		Server:   host.api.Server(name),
		Model:    host.DB.NewModel(name),
		Name:     name,
		Size:     size,
		Location: location,
		Stdin:    os.Stdin,
		Stdout:   os.Stdout,
	}
	return &s, host.DB.Query(`
	
	  INSERT INTO servers (id, name, size, location)
		VALUES (?, ?, ?, ?)
		RETURNING created_at, updated_at
	
	`, s.ID, s.Name, s.Size, s.Location).Scan(&s.CreatedAt, &s.UpdatedAt)
}

func (host *CongoHost) GetServer(id string) (*RemoteServer, error) {
	s := RemoteServer{host: host, Model: host.DB.Model(), Server: host.api.Server(id)}
	return &s, host.DB.Query(`

		SELECT id, name, size, location, created_at, updated_at
		FROM servers
		WHERE id = ?

	`, id).Scan(&s.ID, &s.Name, &s.Size, &s.Location, &s.CreatedAt, &s.UpdatedAt)
}

func (host *CongoHost) ListServers() (servers []*RemoteServer, err error) {
	return servers, host.DB.Query(`

		SELECT id, name, size, location, created_at, updated_at
		FROM servers
		ORDER BY created_at DESC

	`).All(func(scan congo.Scanner) error {
		s := RemoteServer{host: host, Model: host.DB.Model()}
		err = scan(&s.ID, &s.Name, &s.Size, &s.Location, &s.CreatedAt, &s.UpdatedAt)
		servers = append(servers, &s)
		s.Server = host.api.Server(s.ID)
		return err
	})
}

func (s *RemoteServer) Save() error {
	return s.host.DB.Query(`

		UPDATE servers
		SET name = ?,
				size = ?,
				location = ?,
				updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		RETURNING updated_at

	`, s.Name, s.Size, s.Location, s.ID).Scan(&s.UpdatedAt)
}

//go:embed resources/server/prepare-server.sh
var prepareServer string

func (server *RemoteServer) Prepare() error {
	_, stderr, err := server.Run(server.Stdin, fmt.Sprintf(prepareServer, server.Name+"-data"))
	return errors.Wrap(err, stderr.String())
}

func (server *RemoteServer) Deploy(source string) error {
	if _, _, err := server.Copy(source, "/root/congo"); err != nil {
		return errors.Wrap(err, "failed to copy source to server")
	}

	volume, ok := map[string]int64{"SM": 5, "MD": 25, "LG": 50}[server.Size]
	if !ok {
		return errors.New("invalid size")
	}

	err := server.Create(server.Size, server.Location, volume)
	return errors.Wrap(err, "failed to start server")
}

//go:embed resources/server/start-server.sh
var startServer string

func (server *RemoteServer) Restart() error {
	_, out, err := server.Run(nil, fmt.Sprintf(startServer, 8080))
	return errors.Wrap(err, out.String())
}
