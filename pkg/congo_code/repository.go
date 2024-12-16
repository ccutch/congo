package congo_code

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/ccutch/congo/pkg/congo_auth"
	"github.com/sosedoff/gitkit"
	"golang.org/x/crypto/bcrypt"
)

type RepoHost struct {
	auth *congo_auth.Directory
	git  *gitkit.Server
}

func NewRepo(auth *congo_auth.Directory, root string) *RepoHost {
	if err := os.MkdirAll(root, os.ModePerm); err != nil {
		log.Fatal(err)
	}
	repos := RepoHost{auth, gitkit.New(gitkit.Config{
		Dir:        root,
		AutoCreate: true,
		Auth:       true,
		Hooks: &gitkit.HookScripts{
			PostReceive: "curl -X POST http://localhost:8080/hooks/post-receive?repo-id=$1 -d @-",
		},
	})}
	repos.git.AuthFunc = repos.authenticate
	if err := repos.git.Setup(); err != nil {
		log.Fatal(err)
	}
	return &repos
}

func (repos *RepoHost) authenticate(cred gitkit.Credential, req *gitkit.Request) (bool, error) {
	if cred.Username == "" || cred.Password == "" {
		return false, nil
	}

	i, err := repos.auth.Lookup(cred.Username)
	if err != nil {
		return false, err
	}

	err = bcrypt.CompareHashAndPassword(i.PassHash, []byte(cred.Password))
	if err != nil {
		return false, err
	}

	return true, nil
}

func (repos *RepoHost) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, ".git") {
		repos.git.ServeHTTP(w, r)
		return
	}
	http.NotFound(w, r)
}
