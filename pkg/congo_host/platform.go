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
	Launch(region, size string, storage int64) error
	Delete(force bool) error
	Reload() error
	Run(io.Reader, io.Writer, ...string) error
	Copy(source, dest string) (stdout, stderr bytes.Buffer, _ error)
	Assign(domain *Domain) error
	Verify(admin string, domains ...*Domain) error
	Remove(domain *Domain) error
}
