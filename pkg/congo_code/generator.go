package congo_code

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed all:templates/*.tmpl all:templates/**/*.tmpl
var templateFiles embed.FS

//go:embed all:templates/**/*.html all:templates/**/**/*.html all:templates/**/*.sql
var nonTemplateFiles embed.FS

func GenerateExample(dest, name string) error {
	// Step 1: Ensure the destination directory exists
	if err := createDirectory(dest); err != nil {
		return err
	}

	// Step 2: Process template files
	if err := processTemplateFiles(dest, name); err != nil {
		return err
	}

	// Step 3: Copy non-template files
	if err := copyNonTemplateFiles(dest); err != nil {
		return err
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

// processTemplateFiles processes template files and interpolates variables into them
func processTemplateFiles(dest, name string) error {
	return fs.WalkDir(templateFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error walking template files: %w", err)
		}
		if d.IsDir() {
			return nil // Skip directories
		}

		// Read and parse the template
		tmplContent, err := templateFiles.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read template file %s: %w", path, err)
		}

		tmpl, err := template.New(path).Parse(string(tmplContent))
		if err != nil {
			return fmt.Errorf("failed to parse template file %s: %w", path, err)
		}

		// Prepare output path and directory
		outputPath, err := prepareOutputPath(dest, path, true) // Remove .tmpl extension
		if err != nil {
			return err
		}

		// Write the processed template to the destination
		return writeTemplateToFile(tmpl, outputPath, name)
	})
}

// copyNonTemplateFiles copies non-template files directly into the destination
func copyNonTemplateFiles(dest string) error {
	return fs.WalkDir(nonTemplateFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error walking non-template files: %w", err)
		}
		if d.IsDir() {
			return nil // Skip directories
		}

		// Prepare output path and directory
		outputPath, err := prepareOutputPath(dest, path, false)
		if err != nil {
			return err
		}

		// Copy file content
		return copyFile(nonTemplateFiles, path, outputPath)
	})
}

// prepareOutputPath prepares the output file path and ensures its directory exists
func prepareOutputPath(dest, sourcePath string, removeTmplExtension bool) (string, error) {
	relPath, _ := filepath.Rel("templates", sourcePath)
	if removeTmplExtension && strings.HasSuffix(relPath, ".tmpl") {
		relPath = strings.TrimSuffix(relPath, ".tmpl")
	}
	outputPath := filepath.Join(dest, relPath)
	outputDir := filepath.Dir(outputPath)

	// Ensure the output directory exists
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create output directory for %s: %w", outputPath, err)
	}

	return outputPath, nil
}

// writeTemplateToFile writes the parsed template to the specified file
func writeTemplateToFile(tmpl *template.Template, outputPath string, data interface{}) error {
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", outputPath, err)
	}
	defer outputFile.Close()

	if err := tmpl.Execute(outputFile, data); err != nil {
		return fmt.Errorf("failed to execute template for %s: %w", outputPath, err)
	}

	return nil
}

// copyFile copies the content of a file from the source embedded FS to the destination path
func copyFile(sourceFS embed.FS, sourcePath, destPath string) error {
	srcFile, err := sourceFS.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", sourcePath, err)
	}
	defer srcFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", destPath, err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file from %s to %s: %w", sourcePath, destPath, err)
	}

	return nil
}
