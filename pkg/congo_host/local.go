package congo_host

import (
	"bytes"
	_ "embed"
	"io"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

type Target interface {
	SetStdin(io.Reader)
	SetStdout(io.Writer)
	Run(...string) error
}

type LocalHost struct {
	host   *CongoHost
	stdin  io.Reader
	stdout io.Writer
}

func (host *CongoHost) Local() *LocalHost {
	return &LocalHost{host, os.Stdin, os.Stdout}
}

func (h *LocalHost) SetStdin(stdin io.Reader) {
	h.stdin = stdin
}

func (h *LocalHost) SetStdout(stdout io.Writer) {
	h.stdout = stdout
}

func (h *LocalHost) Run(args ...string) error {
	cmd := exec.Command("bash", append([]string{"-c"}, args...)...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = h.stdout
	cmd.Stdin = h.stdin
	return errors.Wrap(cmd.Run(), stderr.String())
}
