package congo_code

import (
	"errors"
	"log"
	"path/filepath"
	"sort"
	"strings"
)

type Repository struct {
	code *CongoCode
	ID   string
	Name string
}

func (code *CongoCode) Repo(id string, opts ...RepoOpt) (*Repository, error) {
	repo := Repository{code, id, id}
	for _, opt := range opts {
		if err := opt(&repo); err != nil {
			return nil, err
		}
	}
	return &repo, nil
}

type RepoOpt func(*Repository) error

func WithName(name string) RepoOpt {
	return func(r *Repository) error {
		r.Name = name
		return nil
	}
}

func (repo *Repository) Path() string {
	return filepath.Join(repo.code.DB.Root, "repos", repo.ID)
}

func (repo *Repository) Open(branch, path string) (*Blob, error) {
	isDir, err := repo.isDir(branch, path)
	return &Blob{repo, isDir, err == nil, branch, path}, err
}

func (repo *Repository) Blobs(branch, path string) (blobs []*Blob, err error) {
	stdout, stderr, err := repo.Run("ls-tree", branch, filepath.Join(".", path)+"/")
	if err != nil {
		log.Println("failed to run git command", err)
		return nil, errors.New(stderr.String())
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
		return nil, errors.New(stderr.String())
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
		return nil, errors.New(stderr.String())
	}
	return strings.Split(strings.TrimSpace(stdout.String()), "\n"), nil
}

func (repo *Repository) isDir(branch, path string) (bool, error) {
	if path == "" || path == "." {
		return true, nil
	}
	stdout, stderr, err := repo.Run("ls-tree", branch, filepath.Join(".", path))
	if err != nil {
		return false, errors.New(stderr.String())
	}
	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return false, errors.New("no such file or directory")
	}
	parts := strings.Fields(output)
	return parts[1] == "tree", nil
}
