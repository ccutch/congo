package congo_code

import (
	"net/http"

	"github.com/ccutch/congo/pkg/congo_auth"
)

type Repository struct {
	code     *CongoCode
	ID, Name string
}

type RepoOpt func(*Repository) error

func (code *CongoCode) Repository(id string, opts ...RepoOpt) (*Repository, error) {
	repo := Repository{code, id, id}
	for _, opt := range opts {
		if err := opt(&repo); err != nil {
			return nil, err
		}
	}

	return &repo, nil
}

func WithName(name string) RepoOpt {
	return func(r *Repository) error {
		r.Name = name
		return nil
	}
}

func (repo *Repository) Serve(auth *congo_auth.Controller, roles ...string) http.Handler {
	return repo.code.Server(auth, roles...)
}
