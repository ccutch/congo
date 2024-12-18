package congo_code

import (
	"fmt"
	"net/http"
)

type Repository struct {
	code     *CongoCode
	ID, Name string
}

func (code *CongoCode) Repo(id string, opts ...RepoOpt) (*Repository, error) {
	repo := Repository{code, id, ""}
	for _, opt := range opts {
		if err := opt(&repo); err != nil {
			return nil, err
		}
	}
	return &repo, nil
}

func (repo *Repository) Server() http.Handler {
	mux := http.NewServeMux()
	mux.Handle(fmt.Sprintf("/%s/", repo.ID), repo.code.git)
	return mux
}

type RepoOpt func(*Repository) error

func WithName(name string) RepoOpt {
	return func(r *Repository) error {
		r.Name = name
		return nil
	}
}
