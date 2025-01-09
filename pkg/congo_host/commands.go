package congo_host

import (
	"bytes"
	_ "embed"
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

//go:embed resources/server/prepare-server.sh
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

	if server.local {
		return errors.New("cannot run on local host")
	}

	_, priKey := server.Keys()
	cmd := exec.Command(
		"ssh",
		"-o", "StrictHostKeyChecking=no",
		"-i", priKey,
		fmt.Sprintf("root@%s", server.IP),
		strings.Join(args, " "),
	)

	var buf bytes.Buffer
	cmd.Stdin = server.Stdin
	cmd.Stdout = server.Stdout
	cmd.Stderr = &buf
	return errors.Wrap(cmd.Run(), buf.String())
}

func (server *Server) run(args ...string) (stdout bytes.Buffer, stderr bytes.Buffer, err error) {
	var cmd *exec.Cmd
	if server.local {
		cmd = exec.Command(args[0], args[1:]...)
	} else {
		_, priKey := server.Keys()
		cmd = exec.Command("ssh",
			"-o", "StrictHostKeyChecking=no",
			"-i", priKey,
			fmt.Sprintf("root@%s", server.IP),
			strings.Join(args, " "),
		)
	}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	return stdout, stderr, cmd.Run()
}

func (server *Server) bash(args ...string) (stdout bytes.Buffer, stderr bytes.Buffer, err error) {
	return server.run(append([]string{"bash", "-c"}, args...)...)
}

func (server *Server) docker(args ...string) (stdout bytes.Buffer, stderr bytes.Buffer, err error) {
	return server.run(append([]string{"docker"}, args...)...)
}

func (server *Server) Copy(source, dest string) error {
	if server.Error != nil {
		return server.Error
	}

	_, priKey := server.Keys()
	cmd := exec.Command(
		"scp",
		"-o", "StrictHostKeyChecking=no",
		"-i", priKey,
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
