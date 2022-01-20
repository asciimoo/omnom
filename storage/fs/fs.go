package fs

import (
	"bytes"
	"compress/gzip"
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
	fname := key + ".gz"
	path := filepath.Join(s.baseDir, "snapshots", getPrefix(key), fname)
	snapshot, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	return snapshot
}

func (s *FSStorage) SaveSnapshot(key string, snapshot []byte) error {
	fname := key + ".gz"
	path := filepath.Join(s.baseDir, "snapshots", getPrefix(key))
	err := mkdir(path)
	if err != nil {
		return err
	}
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write(snapshot)
	w.Close()
	return os.WriteFile(filepath.Join(path, fname), b.Bytes(), 0644)
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
