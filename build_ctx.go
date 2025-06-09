package main

import (
	"archive/tar"
	"bytes"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Embed your Dockerfile
//
//go:embed Dockerfile
//go:embed generate_makefile.py
var embeddedFiles embed.FS

// createBuildContext generates the build context including the current directory and the embedded Dockerfile.
func createBuildContext(baseDir string, dockerfile string, ingoreDirectors []string) (io.Reader, error) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	// Correctly handle paths by ensuring they use Unix-style separators
	fixPathSeparator := func(path string) string {
		return strings.ReplaceAll(path, `\`, `/`) // Replace Windows-style separators
	}

	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip the root directory entry
		if path == baseDir {
			return nil
		}

		// Skip files or directories that start with an underscore
		if strings.HasPrefix(filepath.Base(path), "_") {
			if info.IsDir() {
				return filepath.SkipDir // Skip entire directory if it starts with an underscore
			}
			return nil // Skip this file
		}

		// Skip directories that are in the ignore list
		for _, ignoreDir := range ingoreDirectors {
			if strings.HasPrefix(path, filepath.Join(baseDir, ignoreDir)) {
				if info.IsDir() {
					return filepath.SkipDir // Skip entire directory if it matches an ignore directory
				}
				return nil // Skip this file
			}
		}

		// Ensure file paths use Unix-style separators
		fixedPath := fixPathSeparator(path)

		// Create header
		header, err := tar.FileInfoHeader(info, fixedPath)
		if err != nil {
			return err
		}
		// Adjust header name to be relative to base directory and use correct path separators
		header.Name, err = filepath.Rel(baseDir, path)
		if err != nil {
			return err
		}
		header.Name = fixPathSeparator(header.Name) // Ensure Unix-style separators in the tarball

		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if !info.IsDir() {
			// Write file content
			fmt.Println("Adding file to tar:", fixedPath)
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			if _, err := io.Copy(tw, file); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error adding current directory to tar: %w", err)
	}

	var dockerfileData []byte
	if dockerfile == "" {
		dockerfileData, err = fs.ReadFile(embeddedFiles, "Dockerfile")
		if err != nil {
			return nil, fmt.Errorf("error reading embedded Dockerfile: %w", err)
		}
	} else {
		dockerfileData, err = os.ReadFile(dockerfile)
		if err != nil {
			return nil, fmt.Errorf("error reading custom Dockerfile: %w", err)
		}
		fmt.Println("Using custom dockerfile: ", dockerfile)
	}

	// Add the embedded Dockerfile
	if err := tw.WriteHeader(&tar.Header{
		Name: "Dockerfile",
		Mode: 0600,
		Size: int64(len(dockerfileData)),
	}); err != nil {
		return nil, err
	}
	if _, err := tw.Write(dockerfileData); err != nil {
		return nil, err
	}

	// Add the embedded generate_makefile.py
	genMakeData, err := fs.ReadFile(embeddedFiles, "generate_makefile.py")
	if err != nil {
		return nil, fmt.Errorf("error reading embedded generate_makefile.py: %w", err)
	}
	if err := tw.WriteHeader(&tar.Header{
		Name: "generate_makefile.py",
		Mode: 0600,
		Size: int64(len(genMakeData)),
	}); err != nil {
		return nil, err
	}
	if _, err := tw.Write(genMakeData); err != nil {
		return nil, err
	}

	// Make sure to close the tar writer to flush buffers
	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("error finalizing tarball: %w", err)
	}

	return bytes.NewReader(buf.Bytes()), nil
}
