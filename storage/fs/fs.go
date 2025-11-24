// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

// Package fs implements filesystem-based storage for Omnom snapshots and resources.
//
// This package provides the Storage type which implements the storage.Storage
// interface using the local filesystem. All content is stored in a configurable
// base directory with the following structure:
//
//	<base_dir>/
//	  snapshots/
//	    <2-char-prefix>/
//	      <hash>.html.gz
//	  resources/
//	    <2-char-prefix>/
//	      <hash><extension>
//
// All files are compressed with gzip before being written to disk. The two-character
// prefix directories (based on the first two characters of the content hash) help
// distribute files across multiple directories to avoid filesystem performance issues.
//
// Example usage:
//
//	storage, err := fs.New(config.StorageFilesystem{
//	    RootDir: "./data",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Storage implements the storage.Storage interface
//	err = storage.SaveSnapshot(key, content)
package fs

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/asciimoo/omnom/config"
)

// Storage implements filesystem-based storage for snapshots and resources.
type Storage struct {
	baseDir string
}

// New creates a new filesystem storage backend.
func New(cfg config.StorageFilesystem) (*Storage, error) {
	var err error
	baseDir, err := filepath.Abs(cfg.RootDir)
	if err != nil {
		return nil, err
	}
	if err := mkdir(filepath.Join(baseDir, "snapshots")); err != nil {
		return nil, err
	}
	return &Storage{
		baseDir: baseDir,
	}, nil
}

// FS returns the storage directory as an io/fs.FS for filesystem operations.
func (s *Storage) FS() (fs.FS, error) {
	root, err := os.OpenRoot(s.baseDir)
	if err != nil {
		return nil, err
	}
	return root.FS(), nil
}

// GetSnapshot retrieves a snapshot file by its key.
// Returns nil if the snapshot doesn't exist or cannot be opened.
func (s *Storage) GetSnapshot(key string) io.ReadCloser {
	path := s.getSnapshotPath(key)
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	return f
}

// GetSnapshotSize returns the size in bytes of a stored snapshot file.
// Returns 0 if the snapshot doesn't exist or an error occurs.
func (s *Storage) GetSnapshotSize(key string) uint {
	path := s.getSnapshotPath(key)
	fi, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return uint(fi.Size())
}

// GetResource retrieves a gzip-compressed resource file by its key.
// Returns nil if the resource doesn't exist or cannot be opened.
func (s *Storage) GetResource(key string) io.ReadCloser {
	path := s.getResourcePath(key)
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	return f
}

// GetResourceSize returns the size in bytes of a stored resource file.
// Returns 0 if the resource doesn't exist or an error occurs.
func (s *Storage) GetResourceSize(key string) uint {
	path := s.getResourcePath(key)
	fi, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return uint(fi.Size())
}

// GetResourceURL constructs the HTTP URL path for accessing a resource.
// The URL path includes a two-character prefix directory for distribution.
func (s *Storage) GetResourceURL(key string) string {
	return filepath.Join("/static/data/resources/", getPrefix(key), key)
}

// SaveSnapshot saves a snapshot to disk with gzip compression.
// Creates the necessary directory structure if it doesn't exist.
func (s *Storage) SaveSnapshot(key string, snapshot []byte) error {
	path := s.getSnapshotPath(key)
	err := mkdir(filepath.Dir(path))
	if err != nil {
		return err
	}
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, _ = w.Write(snapshot)
	w.Close()
	return os.WriteFile(path, b.Bytes(), 0600)
}

// SaveResource saves a resource to disk with gzip compression.
// Creates the necessary directory structure if it doesn't exist.
func (s *Storage) SaveResource(key string, resource []byte) error {
	path := s.getResourcePath(key)
	err := mkdir(filepath.Dir(path))
	if err != nil {
		return err
	}
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, _ = w.Write(resource)
	w.Close()
	return os.WriteFile(path, b.Bytes(), 0600)
}

func mkdir(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(dir, os.ModePerm)
		}
		return err
	}
	return nil
}

func (s *Storage) getSnapshotPath(key string) string {
	fname := filepath.Base(key) + ".gz"
	return filepath.Join(s.baseDir, "snapshots", getPrefix(key), fname)
}

func (s *Storage) getResourcePath(key string) string {
	key = filepath.Base(key)
	if len(key) < 32 {
		key = ""
	}
	return filepath.Join(s.baseDir, "resources", getPrefix(key), key)
}

func getPrefix(s string) string {
	if len(s) < 2 {
		return ""
	}
	return s[:2]
}
