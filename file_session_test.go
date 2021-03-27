package session

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var fileSession = &FileSessionHandler{
	Path:        "/tmp/go_session",
	MaxLifeTime: 10 * time.Minute,
}

func TestWrite(t *testing.T) {
	id := randomString(40)
	data := "test"
	err := fileSession.Write(id, []byte(data))
	if err != nil {
		t.Fatalf("error write data: [%s] %v\n", filepath.Join(fileSession.Path, id), err)
	}
}

func TestRead(t *testing.T) {
	id := randomString(40)
	path := filepath.Join(fileSession.Path, id)
	ioutil.WriteFile(path, []byte("test"), os.ModeAppend)
	var data []byte
	data, err := fileSession.Read(id)
	if err != nil {
		t.Fatalf("error read data: [%s] %v\n", id, err)
	}
	if len(data) <= 0 {
		t.Fatalf("error read data: [%s]", id)
	}

	file, _ := os.OpenFile(path, os.O_WRONLY, os.ModeAppend)
	finfo, err := file.Stat()
	finfo.ModTime().AddDate(0, 0, -1)
	id = randomString(40)
	data, err = fileSession.Read(id)
	if err != nil {
		if !os.IsNotExist(err) {
			t.Fatalf("error read data: [%s] %v\n", id, err)
		}
	}
	if len(data) > 0 {
		t.Fatalf("error read data: [%s]\n", id)
	}
}

func TestDestroy(t *testing.T) {
	id := randomString(40)
	path := filepath.Join(fileSession.Path, id)
	ioutil.WriteFile(path, []byte("test"), os.ModeAppend)
	err := fileSession.Destroy(id)
	if err != nil {
		t.Fatalf("cannot destroy session %s: %v", id, err)
	}
	err = fileSession.Destroy("test")
	if !os.IsNotExist(err) {
		t.Fatalf("file should be not exist")
	}
}

func TestGc(t *testing.T) {
	filepath.Walk(fileSession.Path, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		os.Remove(path)
		return nil
	})
	for i := 0; i < 10; i++ {
		id := randomString(40)
		path := filepath.Join(fileSession.Path, id)
		ioutil.WriteFile(path, []byte("test"), os.ModeAppend)
		if i > 4 {
			file, _ := os.OpenFile(path, os.O_WRONLY, os.ModeAppend)
			defer file.Close()
			finfo, _ := file.Stat()
			mtime := finfo.ModTime()
			os.Chtimes(path, mtime.AddDate(0, 0, -1), mtime.AddDate(0, 0, -1))
		}
	}
	err := fileSession.Gc()
	if err != nil {
		t.Fatalf("garbage collect failed: %v", err)
	}
	filepath.Dir(fileSession.Path)
	size := 0
	filepath.Walk(fileSession.Path, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size++
		}
		return nil
	})
	if size != 5 {
		t.Fatalf("garbage collect failed: %d", size)
	}
}
