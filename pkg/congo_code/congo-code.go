package congo_code

import (
	"bytes"
	"log"
	"os/exec"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
	"github.com/sosedoff/gitkit"
	"golang.org/x/crypto/bcrypt"
)

type CongoCode struct {
	app *congo.Application
	git *gitkit.Server
}

func InitCongoCode(app *congo.Application, opts ...CongoCodeOpt) *CongoCode {
	code := CongoCode{app: app}
	for _, opt := range opts {
		if err := opt(&code); err != nil {
			log.Fatal("Failed to setup Congo Code: ", err)
		}
	}
	return &code
}

func (code *CongoCode) run(args ...string) (stdout, stderr bytes.Buffer, _ error) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	return stdout, stderr, cmd.Run()
}

func (code *CongoCode) bash(args ...string) (bytes.Buffer, bytes.Buffer, error) {
	return code.run(append([]string{"bash", "-c"}, args...)...)
}

func (code *CongoCode) docker(args ...string) (bytes.Buffer, bytes.Buffer, error) {
	return code.run(append([]string{"docker"}, args...)...)
}

type CongoCodeOpt func(*CongoCode) error

func WithGitServer(auth *congo_auth.CongoAuth) CongoCodeOpt {
	return func(code *CongoCode) error {
		code.git = gitkit.New(gitkit.Config{
			Dir:        code.app.DB.Root,
			AutoCreate: true,
			Auth:       auth != nil,
		})
		// If auth is provided then we will authenticate with basic auth
		if auth != nil {
			code.git.AuthFunc =
				func(cred gitkit.Credential, req *gitkit.Request) (bool, error) {
					if cred.Username == "" || cred.Password == "" {
						return false, nil
					}
					i, err := auth.Lookup(cred.Username)
					if err != nil {
						return false, err
					}
					pass := []byte(cred.Password)
					err = bcrypt.CompareHashAndPassword(i.PassHash, pass)
					if err != nil {
						return false, err
					}
					return true, nil
				}
		}
		return code.git.Setup()
	}
}
