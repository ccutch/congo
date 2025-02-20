package congo_code

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/ccutch/congo/pkg/congo"
	"github.com/pkg/errors"
)

type Repository struct {
	code *CongoCode
	congo.Model
	Name string
}

func (code *CongoCode) Repositories() ([]*Repository, error) {
	repos := []*Repository{}
	return repos, code.db.Query(`
	
		SELECT id, name, created_at, updated_at
		FROM repositories
		ORDER BY created_at DESC

	`).All(func(scan congo.Scanner) error {
		r := Repository{Model: code.db.Model()}
		repos = append(repos, &r)
		return scan(&r.ID, &r.Name, &r.CreatedAt, &r.UpdatedAt)
	})
}

func (code *CongoCode) NewRepo(id string, opts ...RepoOpt) (*Repository, error) {
	repo := Repository{code, code.db.NewModel(id), id}
	for _, opt := range opts {
		if err := opt(&repo); err != nil {
			return nil, err
		}
	}
	return &repo, code.db.Query(`

		INSERT INTO repositories (id, name)
		VALUES (?, ?)
		RETURNING created_at, updated_at

	`, repo.ID, repo.Name).Scan(&repo.CreatedAt, &repo.UpdatedAt)
}

func (code *CongoCode) GetRepository(id string) (*Repository, error) {
	repo := Repository{Model: code.db.Model()}
	return &repo, code.db.Query(`
	
		SELECT id, name, created_at, updated_at
		FROM repositories
		WHERE id = ?

	`, id).Scan(&repo.ID, &repo.Name, &repo.CreatedAt, &repo.UpdatedAt)
}

func (repo *Repository) Save() error {
	return repo.DB.Query(`
	
		UPDATE repositories
		SET name = ?
		WHERE id = ?
		RETURNING created_at, updated_at

	`, repo.Name, repo.ID).Scan(&repo.CreatedAt, &repo.UpdatedAt)
}

func (repo *Repository) Delete() error {
	return repo.DB.Query(`

		DELETE FROM repositories
		WHERE id = ?

	`, repo.ID).Exec()
}

type RepoOpt func(*Repository) error

func WithName(name string) RepoOpt {
	return func(r *Repository) error {
		r.Name = name
		return nil
	}
}

func (repo *Repository) Path() string {
	return filepath.Join(repo.code.db.Root, "repos", repo.ID)
}

func (repo *Repository) Open(branch, path string) (*Blob, error) {
	isDir, err := repo.isDir(branch, path)
	return &Blob{repo, isDir, err == nil, branch, path}, err
}

func (repo *Repository) Blobs(branch, path string) (blobs []*Blob, err error) {
	stdout, stderr, err := repo.Run("ls-tree", branch, filepath.Join(".", path)+"/")
	if err != nil {
		return nil, errors.Wrap(err, stderr.String())
	}
	for _, line := range strings.Split(strings.TrimSpace(stdout.String()), "\n") {
		if parts := strings.Fields(line); len(parts) >= 4 {
			blobs = append(blobs, &Blob{
				repo:   repo,
				Branch: branch,
				Exists: true,
				isDir:  parts[1] == "tree",
				Path:   parts[3],
			})
		}
	}
	sort.Slice(blobs, func(i, j int) bool {
		if blobs[i].isDir && !blobs[j].isDir {
			return true
		}
		if !blobs[i].isDir && blobs[j].isDir {
			return false
		}
		return blobs[i].Path < blobs[j].Path
	})
	return blobs, nil
}

func (repo *Repository) Branches() ([]string, error) {
	stdout, stderr, err := repo.Run("branch", "--list")
	if err != nil {
		return nil, errors.Wrap(err, stderr.String())
	}
	var branches []string
	for _, line := range strings.Split(strings.TrimSpace(stdout.String()), "\n") {
		if branch := strings.TrimSpace(strings.TrimPrefix(line, "*")); branch != "" {
			branches = append(branches, branch)
		}
	}
	sort.Strings(branches)
	return branches, nil
}

func (repo *Repository) Commits(branch string) ([]string, error) {
	stdout, stderr, err := repo.Run("log", branch, "--pretty=format:%h - %s", "--no-merges")
	if err != nil {
		return nil, errors.Wrap(err, stderr.String())
	}
	return strings.Split(strings.TrimSpace(stdout.String()), "\n"), nil
}

func (repo *Repository) isDir(branch, path string) (bool, error) {
	if path == "" || path == "." {
		return true, nil
	}
	stdout, stderr, err := repo.Run("ls-tree", branch, filepath.Join(".", path))
	if err != nil {
		return false, errors.Wrap(err, stderr.String())
	}
	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return false, errors.New("no such file or directory")
	}
	parts := strings.Fields(output)
	return parts[1] == "tree", nil
}
