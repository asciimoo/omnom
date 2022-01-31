package storage

import (
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/asciimoo/omnom/storage/fs"
)

type Storage interface {
	Init(string) error
	GetSnapshot(string) []byte
	GetSnapshotSize(string) int64
	SaveSnapshot(string, []byte) error
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

func GetSnapshot(key string) ([]byte, error) {
	if store == nil {
		return nil, errors.New("Uninitialized storage")
	}
	return store.GetSnapshot(key), nil
}

func SaveSnapshot(key string, snapshot []byte) error {
	if store == nil {
		return errors.New("Uninitialized storage")
	}
	return store.SaveSnapshot(key, snapshot)
}

func Hash(x []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(x))
}
