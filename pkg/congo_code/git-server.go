package congo_code

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ccutch/congo/pkg/congo_auth"
	"github.com/sosedoff/gitkit"
	"golang.org/x/crypto/bcrypt"
)

func (code *CongoCode) GitServer(auth *congo_auth.Controller, roles ...string) http.Handler {
	git := gitkit.New(gitkit.Config{
		Dir:        filepath.Join(code.db.Root, "repos"),
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
					return false, fmt.Errorf("%s is not a %s", i.Name, role_list)
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

func (repo *Repository) Serve(auth *congo_auth.Controller, roles ...string) http.Handler {
	return repo.code.GitServer(auth, roles...)
}
