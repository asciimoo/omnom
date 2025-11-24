// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

// Package storage provides an abstraction layer for storing snapshots and resources.
//
// This package defines the Storage interface which can be implemented by various
// backends (filesystem, S3, etc.). Currently, only filesystem storage is implemented.
//
// Snapshots are compressed HTML archives of bookmarked web pages, while resources
// are embedded assets like images, stylesheets, and scripts extracted from pages.
//
// All stored content is compressed with gzip to save disk space. Files are organized
// in a two-character prefix directory structure based on their hash to avoid
// filesystem limitations with too many files in a single directory.
//
// The package provides a global storage instance that is initialized at startup
// and used throughout the application. All functions panic if storage is accessed
// before initialization.
//
// Example usage:
//
//	err := storage.Init(cfg.Storage)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Save a snapshot
//	key := storage.Hash(htmlContent) + ".html"
//	err = storage.SaveSnapshot(key, htmlContent)
//
//	// Retrieve a snapshot
//	reader, err := storage.GetSnapshot(key)
//	if err != nil {
//	    return err
//	}
//	defer reader.Close()
package storage

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	iofs "io/fs"

	"github.com/asciimoo/omnom/config"
	"github.com/asciimoo/omnom/storage/fs"
)

// Storage defines the interface for snapshot and resource storage backends.
type Storage interface {
	FS() (iofs.FS, error)
	GetSnapshot(string) io.ReadCloser
	GetSnapshotSize(string) uint
	SaveSnapshot(string, []byte) error
	SaveResource(string, []byte) error
	GetResource(string) io.ReadCloser
	GetResourceSize(string) uint
	GetResourceURL(string) string
}

// ErrUninitialized is returned when storage is accessed before initialization.
var ErrUninitialized = errors.New("uninitialized storage")

// ErrUnknownStorage is returned when an unknown storage type is configured.
var ErrUnknownStorage = errors.New("unknown storage type")

// ErrSnapshotNotFound is returned when a snapshot cannot be found.
var ErrSnapshotNotFound = errors.New("snapshot not found")

// ErrResourceNotFound is returned when a resource cannot be found.
var ErrResourceNotFound = errors.New("resource not found")

var store Storage

func initStorage(sCfg config.Storage) (Storage, error) {
	if sCfg.Filesystem != nil {
		return fs.New(*sCfg.Filesystem)
	}
	return nil, ErrUnknownStorage
}

// Init initializes the storage backend with the given configuration.
// This must be called before using any other storage functions.
// Returns an error if the storage configuration is invalid or initialization fails.
func Init(sCfg config.Storage) error {
	s, err := initStorage(sCfg)
	if err != nil {
		return err
	}
	store = s
	return nil
}

// FS returns the storage backend as an io/fs.FS interface for file system operations.
// This allows reading stored files using the standard library's fs.FS methods.
// Panics if storage has not been initialized.
func FS() iofs.FS {
	if store == nil {
		panic(ErrUninitialized)
	}
	storeFS, err := store.FS()
	if err != nil {
		panic(err)
	}
	return storeFS
}

// GetSnapshot retrieves a snapshot by its key and returns a reader for its contents.
// The returned reader must be closed by the caller when done.
// Returns ErrUninitialized if storage is not initialized, or ErrSnapshotNotFound
// if the snapshot doesn't exist.
func GetSnapshot(key string) (io.ReadCloser, error) {
	if store == nil {
		return nil, ErrUninitialized
	}
	r := store.GetSnapshot(key)
	if r == nil {
		return nil, ErrSnapshotNotFound
	}
	return r, nil
}

// GetResource retrieves a resource by its key and returns a reader for its contents.
// Resources are typically images, stylesheets, or other embedded assets from web pages.
// The returned reader must be closed by the caller when done.
// Returns ErrUninitialized if storage is not initialized, or ErrResourceNotFound
// if the resource doesn't exist.
func GetResource(key string) (io.ReadCloser, error) {
	if store == nil {
		return nil, ErrUninitialized
	}
	r := store.GetResource(key)
	if r == nil {
		return nil, ErrResourceNotFound
	}
	return r, nil
}

// SaveSnapshot stores a snapshot with the given key.
// The snapshot data is compressed before storage to save disk space.
// Returns ErrUninitialized if storage is not initialized, or an error if saving fails.
func SaveSnapshot(key string, snapshot []byte) error {
	if store == nil {
		return ErrUninitialized
	}
	return store.SaveSnapshot(key, snapshot)
}

// SaveResource stores a resource with the given key.
// Resources are typically images, stylesheets, or other embedded assets.
// The resource data is compressed before storage to save disk space.
// Returns ErrUninitialized if storage is not initialized, or an error if saving fails.
func SaveResource(key string, resource []byte) error {
	if store == nil {
		return ErrUninitialized
	}
	return store.SaveResource(key, resource)
}

// GetSnapshotSize returns the size in bytes of a stored snapshot.
// Returns 0 if the snapshot doesn't exist.
// Panics if storage has not been initialized.
func GetSnapshotSize(key string) uint {
	if store == nil {
		panic("Uninitialized storage")
	}
	return store.GetSnapshotSize(key)
}

// GetResourceSize returns the size in bytes of a stored resource.
// Returns 0 if the resource doesn't exist.
// Panics if storage has not been initialized.
func GetResourceSize(key string) uint {
	if store == nil {
		panic("Uninitialized storage")
	}
	return store.GetResourceSize(key)
}

// GetResourceURL returns the URL path for accessing a resource via HTTP.
// This is typically used to generate URLs for embedded images and assets in HTML.
// Panics if storage has not been initialized.
func GetResourceURL(key string) string {
	if store == nil {
		panic("Uninitialized storage")
	}
	return store.GetResourceURL(key)
}

// Hash computes a SHA256 hash of the given bytes and returns it as a hexadecimal string.
// This is used to generate unique keys for storing snapshots and resources.
func Hash(x []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(x))
}
