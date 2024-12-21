package congo_code

import (
	"bytes"
	"cmp"
	"errors"
	"io/fs"
	"log"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (repo *Repository) NewClient(branch string) *GitClient {
	root := filepath.Join(repo.code.root, "repos")
	return &GitClient{repo, root, cmp.Or(branch, "HEAD")}
}

type GitClient struct {
	repo   *Repository
	root   string
	branch string
}

func (git *GitClient) path() string {
	return filepath.Join(git.root, "repos", git.repo.ID+".git")
}

func (git *GitClient) run(args ...string) (string, error) {
	cmd, buf := exec.Command("git", args...), bytes.Buffer{}
	cmd.Dir = git.path()
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (git *GitClient) Init() error {
	_, err := git.run("init", "--bare")
	return err
}

func (git *GitClient) Clone(source string) error {
	_, err := git.run("clone", "--bare", source, git.path())
	return err
}

func (git *GitClient) Copy(dest string) error {
	_, err := git.run("clone", git.path(), dest)
	return err
}

func (git *GitClient) LsBranches() (branches []string, err error) {
	stdout, err := git.run("branch", "--list")
	if err != nil {
		return nil, err
	}
	for _, line := range strings.Split(strings.TrimSpace(stdout), "\n") {
		if branch := strings.TrimSpace(strings.TrimPrefix(line, "*")); branch != "" {
			branches = append(branches, branch)
		}
	}
	sort.Strings(branches)
	return branches, nil
}

func (git *GitClient) LsCommits() ([]string, error) {
	stdout, err := git.run("log", git.branch, "--pretty=format:%h - %s", "--no-merges")
	return strings.Split(strings.TrimSpace(stdout), "\n"), err
}

func (git *GitClient) IsDir(path string) (bool, error) {
	if path == "" || path == "." {
		return true, nil
	}
	stdout, err := git.run("ls-tree", git.branch, filepath.Join(".", path))
	if err != nil {
		return false, err
	}
	if stdout = strings.TrimSpace(stdout); stdout == "" {
		return false, errors.New("no such file or directory")
	}
	parts := strings.Fields(stdout)
	return parts[1] == "tree", nil
}

func (git *GitClient) LsTree(path string) (blobs []*Blob) {
	stdout, err := git.run("ls-tree", git.branch, filepath.Join(".", path)+"/")
	if err != nil {
		log.Println("failed to run git command", err)
		return nil
	}
	for _, line := range strings.Split(strings.TrimSpace(stdout), "\n") {
		if parts := strings.Fields(line); len(parts) >= 4 {
			blobs = append(blobs, &Blob{
				GitClient: git,
				Exists:    true,
				isDir:     parts[1] == "tree",
				Path:      parts[3],
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
	return blobs
}

func (git *GitClient) Open(path string) (*Blob, error) {
	isDir, err := git.IsDir(path)
	return &Blob{git, isDir, err == nil, path}, err
}

type Blob struct {
	*GitClient
	isDir  bool
	Exists bool
	Path   string
}

func (blob *Blob) Read(p []byte) (int, error) {
	content, err := blob.Content()
	if err != nil {
		return 0, err
	}
	return strings.NewReader(content).Read(p)
}
func (blob *Blob) Close() error {
	return nil
}
func (blob *Blob) Stat() (fs.FileInfo, error) {
	return blob, nil
}
func (blob *Blob) Content() (string, error) {
	return blob.run("show", blob.branch+":"+blob.Path)
}
func (blob *Blob) Lines() ([]string, error) {
	content, err := blob.Content()
	return strings.Split(content, "\n"), err
}
func (blob *Blob) Files() []*Blob {
	return blob.GitClient.LsTree(blob.Path)
}
func (blob *Blob) Dir() string {
	dir := filepath.Dir(blob.Path)
	if dir == "." {
		return ""
	}
	return dir
}
func (blob *Blob) Name() string {
	return filepath.Base(blob.Path)
}
func (blob *Blob) Size() int64 {
	sizeStr, err := blob.run("cat-file", "-s", blob.branch+":"+blob.Path)
	if err != nil {
		return 0 // Handle error appropriately
	}
	size, _ := strconv.ParseInt(strings.TrimSpace(sizeStr), 10, 64)
	return size
}
func (blob *Blob) Mode() fs.FileMode {
	if blob.isDir {
		return fs.ModeDir
	}
	return 0
}
func (blob *Blob) ModTime() time.Time { return time.Now() }
func (blob *Blob) IsDir() bool        { return blob.isDir }
func (*Blob) Sys() interface{}        { return nil }
