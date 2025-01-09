package congo_host

import (
	"bytes"
	_ "embed"
)

func (server *RemoteServer) bash(args ...string) (stdout bytes.Buffer, stderr bytes.Buffer, err error) {
	return server.Run(nil, append([]string{"bash", "-c"}, args...)...)
}

func (server *RemoteServer) docker(args ...string) (stdout bytes.Buffer, stderr bytes.Buffer, err error) {
	return server.Run(nil, append([]string{"docker"}, args...)...)
}
