package congo_host

import (
	"bytes"
	"io"
)

type Platform interface {
	Init(*CongoHost)
	Server(name string) Server
}

type Server interface {
	Addr() string
	Create(region, size string, storage int64) error
	Delete(purge, force bool) error
	Reload() error
	Run(io.Reader, io.Writer, ...string) error
	Copy(source, dest string) (stdout, stderr bytes.Buffer, _ error)
}

