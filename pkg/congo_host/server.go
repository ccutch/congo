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
	SetStdin(io.Reader)
	SetStdout(io.Writer)
	Run(...string) error
}

type LocalServer struct {
	host   *CongoHost
	stdin  io.Reader
	stdout io.Writer
}

func (host *CongoHost) Local() *LocalServer {
	return &LocalServer{host, os.Stdin, os.Stdout}
}

func (s *LocalServer) SetStdin(stdin io.Reader) {
	s.stdin = stdin
}

func (s *LocalServer) SetStdout(stdout io.Writer) {
	s.stdout = stdout
}

func (s *LocalServer) Run(args ...string) error {
	cmd := exec.Command("bash", append([]string{"-c"}, args...)...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = s.stdout
	cmd.Stdin = s.stdin
	return errors.Wrap(cmd.Run(), stderr.String())
}

type RemoteServer struct {
	Server
	congo.Model
	host     *CongoHost
	Name     string
	Size     string
	Location string

	Stdin  io.Reader
	Stdout io.Writer
}

func (host *CongoHost) NewServer(name, size, location string) (*RemoteServer, error) {
	s := RemoteServer{
		Server:   host.api.Server(name),
		Model:    host.DB.NewModel(name),
		host:     host,
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
	s := RemoteServer{Model: host.DB.Model(), Server: host.api.Server(id), host: host, Stdin: os.Stdin, Stdout: os.Stdout}
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
		s := RemoteServer{Model: host.DB.Model(), host: host, Stdin: os.Stdin, Stdout: os.Stdout}
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

func (s *RemoteServer) Delete(purge, force bool) error {
	if err := s.Server.Delete(purge, force); !force && err != nil {
		return errors.Wrap(err, "failed to delete server")
	}
	return s.host.DB.Query(`

		DELETE FROM servers
		WHERE id = ? 
	
	`, s.ID).Exec()
}

func (s *RemoteServer) SetStdin(stdin io.Reader) {
	s.Stdin = stdin
}

func (s *RemoteServer) SetStdout(stdout io.Writer) {
	s.Stdout = stdout
}

func (s *RemoteServer) Run(args ...string) error {
	return s.Server.Run(s.Stdin, s.Stdout, args...)
}

//go:embed resources/server/prepare-server.sh
var prepareServer string

func (server *RemoteServer) Prepare() error {
	return server.Run(fmt.Sprintf(prepareServer, server.Name+"-data"))
}

func (server *RemoteServer) Deploy(source string) error {
	if _, _, err := server.Copy(source, "/root/congo"); err != nil {
		return errors.Wrap(err, "failed to copy source to server")
	}

	return server.Restart()
}

//go:embed resources/server/start-server.sh
var startServer string

func (server *RemoteServer) Restart() error {
	return server.Run(fmt.Sprintf(startServer, 8080))
}
