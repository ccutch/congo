package congo_host

import (
	_ "embed"
	"fmt"
)

//go:embed resources/generate-certs.sh
var generateCerts string

func (server *Server) GenerateCerts(domain string) {
	if server.Err != nil {
		return
	}
	server.Err = server.Run(fmt.Sprintf(generateCerts, domain))
}
