package congo_host

import (
	"bytes"
	_ "embed"
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

//go:embed resources/prepare-server.sh
var prepareServer string

func (server *Server) prepareServer() {
	if server.Error != nil {
		return
	}

	server.Error = server.Run(fmt.Sprintf(prepareServer, server.Name+"-data"))
}

func (server *Server) Run(args ...string) error {
	if server.Error != nil {
		return server.Error
	}

	cmd := exec.Command(
		"ssh",
		"-o", "StrictHostKeyChecking=no",
		"-i", server.priKey,
		fmt.Sprintf("root@%s", server.IP),
		strings.Join(args, " "),
	)

	var buf bytes.Buffer
	cmd.Stdin = server.Stdin
	cmd.Stdout = server.Stdout
	cmd.Stderr = &buf
	return errors.Wrap(cmd.Run(), buf.String())
}

func (server *Server) Copy(source, dest string) error {
	if server.Error != nil {
		return server.Error
	}

	cmd := exec.Command(
		"scp",
		"-o", "StrictHostKeyChecking=no",
		"-i", server.priKey,
		source,
		fmt.Sprintf("root@%s:%s", server.IP, dest),
	)

	var buf bytes.Buffer
	cmd.Stdin = server.Stdin
	cmd.Stdout = server.Stdout
	cmd.Stderr = &buf
	return errors.Wrap(cmd.Run(), buf.String())
}

func (server *Server) Deploy(source string) error {
	if server.Error == nil {
		server.Error = server.Copy(source, "/root/congo")
	}

	if server.Error == nil {
		server.Start()
	}

	return server.Error
}
