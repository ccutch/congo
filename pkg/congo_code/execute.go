package congo_code

import (
	"bytes"
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

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

func (repo *Repository) Build(branch, path string) (string, error) {
	dir, err := os.MkdirTemp("", "congo-build-*")
	if err != nil {
		return "", err
	}

	log.Println("dir", dir)
	if err = repo.Copy(dir); err != nil {
		return "", err
	}

	cmd := exec.Command("/usr/local/go/bin/go", "build", "-o", filepath.Join(dir, "congo"), path)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", errors.New("Build Failed: " + err.Error() + " " + string(output))
	}

	return filepath.Join(dir, "congo"), nil
}
