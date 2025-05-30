// SPDX-FileContributor: Adam Tauber <asciimoo@gmail.com>
//
// SPDX-License-Identifier: AGPLv3+

package fs

import (
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"

	"github.com/asciimoo/omnom/config"
)

type FSStorage struct {
	baseDir string
}

func New() *FSStorage {
	return &FSStorage{}
}

func (s *FSStorage) Init(sCfg config.Storage) error {
	var err error
	s.baseDir, err = filepath.Abs(sCfg.RootDir)
	if err != nil {
		return err
	}
	return mkdir(filepath.Join(s.baseDir, "snapshots"))
}

func (s *FSStorage) GetSnapshot(key string) io.ReadCloser {
	path := s.getSnapshotPath(key)
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	return f
}

func (s *FSStorage) GetSnapshotSize(key string) uint {
	path := s.getSnapshotPath(key)
	fi, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return uint(fi.Size())
}

func (s *FSStorage) GetResource(key string) io.ReadCloser {
	path := s.getResourcePath(key)
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	return f
}

func (s *FSStorage) GetResourceSize(key string) uint {
	path := s.getResourcePath(key)
	fi, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return uint(fi.Size())
}

func (s *FSStorage) GetResourceURL(key string) (string, bool) {
	return filepath.Join("/static/data/resources/", getPrefix(key), key), false
}

func (s *FSStorage) SaveSnapshot(key string, snapshot []byte) error {
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

func (s *FSStorage) SaveResource(key string, resource []byte) error {
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

func (s *FSStorage) getSnapshotPath(key string) string {
	fname := filepath.Base(key) + ".gz"
	return filepath.Join(s.baseDir, "snapshots", getPrefix(key), fname)
}

func (s *FSStorage) getResourcePath(key string) string {
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
