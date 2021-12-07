package fs

import (
	"os"
	"path/filepath"
)

type FSStorage struct {
	baseDir string
}

func New() *FSStorage {
	return &FSStorage{}
}

func (s *FSStorage) Init(dir string) error {
	var err error
	s.baseDir, err = filepath.Abs(dir)
	if err != nil {
		return err
	}
	mkdir(filepath.Join(s.baseDir, "snapshots"))
	return nil
}

func (s *FSStorage) GetSnapshot(key string) []byte {
	path := filepath.Join(s.baseDir, "snapshots", getPrefix(key), key)
	snapshot, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return snapshot
}

func (s *FSStorage) SaveSnapshot(key string, snapshot []byte) error {
	path := filepath.Join(s.baseDir, "snapshots", getPrefix(key))
	err := mkdir(path)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(path, key), snapshot, 0644)
}

func mkdir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, os.ModePerm)
	}
	return nil
}

func getPrefix(s string) string {
	if len(s) < 2 {
		return ""
	}
	return s[:2]
}
