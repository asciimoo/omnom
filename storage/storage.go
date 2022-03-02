package storage

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"

	"github.com/asciimoo/omnom/storage/fs"
)

type Storage interface {
	Init(string) error
	GetSnapshot(string) io.ReadCloser
	GetSnapshotSize(string) uint
	SaveSnapshot(string, []byte) error
	SaveResource(string, []byte) error
	GetResource(string) io.ReadCloser
	GetResourceSize(string) uint
}

var store Storage

var storages = map[string]Storage{
	"fs": fs.New(),
}

func Init(sType string, sParams string) error {
	if s, ok := storages[sType]; ok {
		if err := s.Init(sParams); err != nil {
			return err
		}
		store = s
		return nil
	}
	return errors.New("Unknown storage type")
}

func GetSnapshot(key string) (io.ReadCloser, error) {
	if store == nil {
		return nil, errors.New("Uninitialized storage")
	}
	r := store.GetSnapshot(key)
	if r == nil {
		return nil, errors.New("Snapshot not found")
	}
	return r, nil
}

func GetResource(key string) (io.ReadCloser, error) {
	if store == nil {
		return nil, errors.New("Uninitialized storage")
	}
	r := store.GetResource(key)
	if r == nil {
		return nil, errors.New("Resource not found")
	}
	return r, nil
}

func SaveSnapshot(key string, snapshot []byte) error {
	if store == nil {
		return errors.New("Uninitialized storage")
	}
	return store.SaveSnapshot(key, snapshot)
}

func SaveResource(key string, resource []byte) error {
	if store == nil {
		return errors.New("Uninitialized storage")
	}
	return store.SaveResource(key, resource)
}

func GetSnapshotSize(key string) uint {
	if store == nil {
		panic("Uninitialized storage")
	}
	return store.GetSnapshotSize(key)
}

func GetResourceSize(key string) uint {
	if store == nil {
		panic("Uninitialized storage")
	}
	return store.GetResourceSize(key)
}

func Hash(x []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(x))
}
