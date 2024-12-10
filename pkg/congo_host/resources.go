package congo_host

import _ "embed"

//go:embed resources/setup-server.sh
var setupServer string

//go:embed resources/start-server.sh
var startServer string

//go:embed resources/generate-certs.sh
var generateCerts string
