package session

import (
	"bytes"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileSessionHandler file session
type FileSessionHandler struct {
	Path        string
	MaxLifeTime time.Duration
	fileMutex   sync.RWMutex
}

func (f *FileSessionHandler) Open() error {
	return nil
}

func (f *FileSessionHandler) Read(sessionID string) ([]byte, error) {
	f.fileMutex.RLock()
	defer f.fileMutex.RUnlock()
	file, err := os.OpenFile(filepath.Join(f.Path, sessionID), os.O_RDONLY, os.ModeAppend)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if fileInfo.ModTime().Add(f.MaxLifeTime).Before(time.Now()) {
		return nil, nil
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(file)
	return buf.Bytes(), err
}

func (f *FileSessionHandler) Write(sessionID string, p []byte) error {
	f.fileMutex.Lock()
	defer f.fileMutex.Unlock()
	file, err := os.OpenFile(filepath.Join(f.Path, sessionID), os.O_CREATE|os.O_RDWR, os.ModeAppend)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(p)
	return err
}

// Close session
func (f *FileSessionHandler) Close() error {
	return nil
}

// Destroy session
func (f *FileSessionHandler) Destroy(sessionID string) error {
	return os.Remove(filepath.Join(f.Path, sessionID))
}

// Gc Cleanup old sessions
func (f *FileSessionHandler) Gc() error {
	return filepath.Walk(f.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if info.ModTime().Add(f.MaxLifeTime).Before(time.Now()) {
			return os.Remove(path)
		}
		return nil
	})
}
