package congo_code

import (
	"bytes"
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

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
