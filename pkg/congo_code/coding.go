package congo_code

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ccutch/congo/pkg/congo_auth"
	"github.com/sosedoff/gitkit"
	"golang.org/x/crypto/bcrypt"
)

type CongoCode struct {
	root string
}

func InitCongoCode(root string, opts ...CongoCodeOpt) *CongoCode {
	code := CongoCode{root: root}
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

func WithGitServer(auth *congo_auth.Controller) CongoCodeOpt {
	return func(code *CongoCode) error {
		code.GitServer(auth)
		return nil
	}
}

func (code *CongoCode) GitServer(auth *congo_auth.Controller, roles ...string) http.Handler {
	if len(roles) == 0 {
		roles = []string{auth.DefaultRole}
	}

	git := gitkit.New(gitkit.Config{
		Dir:        filepath.Join(code.root, "repos"),
		AutoCreate: true,
		Auth:       auth != nil,
	})

	// If auth is provided then we will authenticate with basic auth
	if auth != nil {
		git.AuthFunc =
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
				if !slices.Contains(roles, i.Role) {
					role_list := strings.Join(roles, " or ")
					return false, fmt.Errorf("%s is not a %s", i.Username, role_list)
				}
				return true, nil
			}
	}

	if err := git.Setup(); err != nil {
		log.Fatalf("Failed to set repository server: %s", err)
		return nil
	}

	return git
}
