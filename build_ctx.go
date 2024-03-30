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
)

// Embed your Dockerfile
//
//go:embed Dockerfile
var embeddedFiles embed.FS

// createBuildContext generates the build context including the current directory and the embedded Dockerfile.
func createBuildContext(baseDir string) (io.Reader, error) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	// Add the current directory to the tar
	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip the root directory entry
		if path == baseDir {
			return nil
		}
		// Create header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		// Ensure header name is relative to base directory
		header.Name, err = filepath.Rel(baseDir, path)
		if err != nil {
			return err
		}
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if !info.IsDir() {
			// Write file content
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

	// Add the embedded Dockerfile
	dockerfileData, err := fs.ReadFile(embeddedFiles, "Dockerfile")
	if err != nil {
		return nil, fmt.Errorf("error reading embedded Dockerfile: %w", err)
	}
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

	// Make sure to close the tar writer to flush buffers
	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("error finalizing tarball: %w", err)
	}

	return bytes.NewReader(buf.Bytes()), nil
}
