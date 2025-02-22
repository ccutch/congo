package congo_code

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Blob struct {
	repo   *Repository
	isDir  bool
	Exists bool
	Branch string
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

func (blob *Blob) Dir() string {
	dir := filepath.Dir(blob.Path)
	if dir == "." {
		return ""
	}

	return dir
}

func (blob *Blob) Files() ([]*Blob, error) {
	return blob.repo.Blobs(blob.Branch, blob.Path)
}

func (blob *Blob) Content() (string, error) {
	stdout, stderr, err := blob.repo.Run("show", blob.Branch+":"+blob.Path)
	if err != nil {
		return "", errors.New(stderr.String())
	}

	return stdout.String(), nil
}

func (blob *Blob) Lines() ([]string, error) {
	content, err := blob.Content()
	return strings.Split(content, "\n"), err
}

func (blob *Blob) Name() string {
	return filepath.Base(blob.Path)
}

func (blob *Blob) Size() int64 {
	stdout, _, err := blob.repo.Run("cat-file", "-s", blob.Branch+":"+blob.Path)
	if err != nil {
		return 0 // Handle error appropriately
	}

	size, _ := strconv.ParseInt(strings.TrimSpace(stdout.String()), 10, 64)
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

func (blob *Blob) FileType() string {
	ext := strings.ToLower(filepath.Ext(blob.Path))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".svg", ".webp":
		return "image/" + ext[1:]
	default:
		return "text/plain"
	}
}

func (blob *Blob) IsImage() bool {
	ext := strings.ToLower(filepath.Ext(blob.Path))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".svg", ".webp":
		return true
	default:
		return false
	}
}
