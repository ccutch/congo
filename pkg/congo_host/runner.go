package congo_host

import "bytes"

type Runner interface {
	Run(args ...string) (stdout, stderr bytes.Buffer, _ error)
}
