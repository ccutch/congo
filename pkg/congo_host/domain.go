package congo_host

import (
	_ "embed"
	"fmt"
)

//go:embed resources/generate-certs.sh
var generateCerts string

func (server *Server) RegisterDomain(domain string) {
	if server.Error != nil {
		return
	}
	server.Error = server.Run(fmt.Sprintf(generateCerts, domain))
}
