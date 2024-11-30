package hosting

import _ "embed"

//go:embed resources/install-golang.sh
var installGolang string

//go:embed resources/mount-volume.sh
var mountVolume string

//go:embed resources/setup-firewall.sh
var setupFirewall string

//go:embed resources/start-server.sh
var startServer string

//go:embed resources/generate-certs.sh
var generateCerts string
