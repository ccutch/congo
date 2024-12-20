package congo_code

import (
	"fmt"
	"net/http"
)

type Repository struct {
	code     *CongoCode
	ID, Name string
	*http.ServeMux
}

type RepoOpt func(*Repository) error

func (code *CongoCode) Repository(id string, opts ...RepoOpt) (*Repository, error) {
	repo := Repository{code, id, id, nil}
	for _, opt := range opts {
		if err := opt(&repo); err != nil {
			return nil, err
		}
	}
	repo.ServeMux = http.NewServeMux()
	repo.ServeMux.Handle(fmt.Sprintf("/%s/", repo.ID), repo.code.git)
	return &repo, nil
}

func WithName(name string) RepoOpt {
	return func(r *Repository) error {
		r.Name = name
		return nil
	}
}
