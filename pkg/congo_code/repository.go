package congo_code

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ccutch/congo/pkg/congo_auth"
)

type Repository struct {
	code     *CongoCode
	ID, Name string
}

func (code *CongoCode) Repository(id string, opts ...RepoOpt) (*Repository, error) {
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

func (repo *Repository) Serve(auth *congo_auth.Controller, roles ...string) http.Handler {
	return repo.code.Server(auth, roles...)
}

func (repo *Repository) Path() string {
	return filepath.Join(repo.code.DB.Root, "repos", repo.ID)
}

func (repo *Repository) Run(args ...string) (stdout bytes.Buffer, stderr bytes.Buffer, err error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = repo.Path()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	return stdout, stderr, cmd.Run()
}

func (repo *Repository) Copy(dest string) error {
	_, stderr, err := repo.Run("clone", repo.Path(), dest)
	if err != nil {
		return errors.New(stderr.String())
	}

	return nil
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

func (repo *Repository) Blobs(branch, path string) ([]*Blob, error) {
	stdout, stderr, err := repo.Run("ls-tree", branch, filepath.Join(".", path)+"/")
	if err != nil {
		log.Println("failed to run git command", err)
		return nil, errors.New(stderr.String())
	}

	var blobs []*Blob
	for _, line := range strings.Split(strings.TrimSpace(stdout.String()), "\n") {
		if parts := strings.Fields(line); len(parts) >= 4 {
			blobs = append(blobs, &Blob{
				Repository: repo,
				Branch:     branch,
				Exists:     true,
				isDir:      parts[1] == "tree",
				Path:       parts[3],
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

func (repo *Repository) Open(branch, path string) (*Blob, error) {
	isDir, err := repo.isDir(branch, path)
	return &Blob{repo, isDir, err == nil, branch, path}, err
}
