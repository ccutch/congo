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
	Create(region, size string, storage int64) error
	Delete(purge, force bool) error
	Reload() error
	Run(_ io.Reader, args ...string) (stdout, stderr bytes.Buffer, _ error)
	Copy(source, dest string) (stdout, stderr bytes.Buffer, _ error)
	Domain(name string) TDomain
}

type TDomain interface {
	Verify(other ...string) error
}
