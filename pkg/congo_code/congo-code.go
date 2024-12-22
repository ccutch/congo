package congo_code

import (
	"bytes"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/ccutch/congo/pkg/congo_auth"
	"github.com/sosedoff/gitkit"
	"golang.org/x/crypto/bcrypt"
)

//go:embed all:migrations
var migrations embed.FS

type CongoCode struct {
	DB *congo.Database
}

func InitCongoCode(root string, opts ...CongoCodeOpt) *CongoCode {
	code := CongoCode{DB: congo.SetupDatabase(root, "code.db", migrations)}
	if err := code.DB.MigrateUp(); err != nil {
		log.Fatal("Failed to setup code db:", err)
	}

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

func (code *CongoCode) Server(auth *congo_auth.Controller, roles ...string) http.Handler {
	if len(roles) == 0 {
		roles = []string{auth.DefaultRole}
	}

	git := gitkit.New(gitkit.Config{
		Dir:        filepath.Join(code.DB.Root, "repos"),
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

				if _, err := code.GetAccessToken(cred.Username, cred.Password); err == nil {
					return true, nil
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

type CongoCodeOpt func(*CongoCode) error
