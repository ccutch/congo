package congo_code

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ccutch/congo/apps"
	"github.com/pkg/errors"
)

func GenerateExample(name, dest, tmpl string) error {
	if err := createDirectory(dest); err != nil {
		return errors.Wrap(err, "failed to create destination directory")
	}
	if err := copySourceFiles(name, dest, tmpl); err != nil {
		return errors.Wrap(err, "failed to copy source files")
	}
	if err := copyResourceFiles(dest, tmpl); err != nil {
		return errors.Wrap(err, "failed to copy resource files")
	}
	return nil
}

// createDirectory ensures the destination directory exists
func createDirectory(dest string) error {
	if err := os.MkdirAll(dest, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}
	return nil
}

// copySourceFiles processes template files and interpolates variables into them
func copySourceFiles(name, dest, tmpl string) error {
	source, err := fs.Sub(apps.SourceFiles, tmpl)
	if err != nil {
		return errors.Wrap(err, "failed to create source filesystem")
	}
	return fs.WalkDir(source, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return errors.Wrap(err, "failed to copy source files")
		}
		if d.IsDir() {
			return nil
		}
		file, err := source.Open(path)
		if err != nil {
			return errors.Wrap(err, "failed to open file")
		}
		defer file.Close()
		content, err := io.ReadAll(file)
		if err != nil {
			return errors.Wrap(err, "failed to read file")
		}
		pkg := fmt.Sprintf("github.com/ccutch/congo/apps/%s", tmpl)
		return copyFile(path, dest, strings.ReplaceAll(string(content), pkg, name))
	})
}

// copyResourceFiles copies non-template files directly into the destination
func copyResourceFiles(dest, tmpl string) error {
	source, err := fs.Sub(apps.ResourceFiles, tmpl)
	if err != nil {
		return errors.Wrap(err, "failed to create source filesystem")
	}
	return fs.WalkDir(source, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return errors.Wrap(err, "failed to copy resource files")
		}
		if d.IsDir() {
			return nil
		}
		file, err := source.Open(path)
		if err != nil {
			return errors.Wrap(err, "failed to open file")
		}
		defer file.Close()
		content, err := io.ReadAll(file)
		if err != nil {
			return errors.Wrap(err, "failed to read file")
		}
		return copyFile(path, dest, string(content))
	})
}

func copyFile(source, dest, content string) error {
	//create directory if it doesn't exist
	dir := filepath.Dir(filepath.Join(dest, source))
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return errors.Wrap(err, "failed to create directory")
		}
	}
	file, err := os.Create(filepath.Join(dest, source))
	if err != nil {
		return errors.Wrap(err, "failed to create destination file")
	}
	defer file.Close()
	_, err = file.WriteString(content)
	return errors.Wrap(err, "failed to write file")
}
