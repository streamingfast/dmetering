package file

import (
	"bytes"
	"io"
	"os"
	"sync"
)

type File interface {
	io.Reader
	io.Writer
	io.Closer
}

type mockFile struct {
	buf *bytes.Buffer
}

func (f *mockFile) Read(p []byte) (n int, err error) {
	return f.buf.Read(p)
}

func (f *mockFile) Write(p []byte) (n int, err error) {
	return f.buf.Write(p)
}

func (f *mockFile) Close() error {
	return nil
}

type FS interface {
	Open(name string) (File, error)
	Create(name string) (File, error)
	Delete(name string) error
}

type OSFS struct{}

func (OSFS) Open(name string) (File, error) {
	return os.Open(name)
}

func (OSFS) Create(name string) (File, error) {
	return os.Create(name)
}

func (OSFS) Delete(name string) error {
	return os.Remove(name)
}

type mockFS struct {
	lock sync.Mutex

	files map[string]File
}

func newMockFS() *mockFS {
	return &mockFS{
		files: make(map[string]File),
	}
}

func (fs *mockFS) Open(name string) (File, error) {
	fs.lock.Lock()
	defer fs.lock.Unlock()

	f, ok := fs.files[name]
	if !ok {
		return nil, os.ErrNotExist
	}
	return f, nil
}

func (fs *mockFS) Create(name string) (File, error) {
	fs.lock.Lock()
	defer fs.lock.Unlock()

	f, ok := fs.files[name]
	if ok {
		return nil, os.ErrExist
	}

	b := new(bytes.Buffer)
	f = &mockFile{buf: b}
	fs.files[name] = f
	return f, nil
}

func (fs *mockFS) Delete(name string) error {
	fs.lock.Lock()
	defer fs.lock.Unlock()

	delete(fs.files, name)
	return nil
}
