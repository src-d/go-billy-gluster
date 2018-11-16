package gluster

import (
	"fmt"
	"io"
	"os"

	"github.com/gluster/gogfapi/gfapi"
	billy "gopkg.in/src-d/go-billy.v4"
)

var (
	_ billy.File = new(File)
)

type mode int

const (
	ErrReadOnly  = "cannot write in %s, the file is read only"
	ErrWriteOnly = "cannot read from %s, the file is write only"

	read  mode = 0
	write mode = 1
)

type File struct {
	path  string
	g     *gfapi.File
	flags int
}

func NewFile(name string, g *gfapi.File, flags int) *File {
	return &File{path: name, g: g, flags: flags}
}

func (f *File) checkFlags(expected mode) error {
	switch expected {
	case read:
		if f.flags&3 == os.O_WRONLY {
			return fmt.Errorf(ErrWriteOnly, f.path)
		}
	case write:
		if f.flags&3 == os.O_RDONLY {
			return fmt.Errorf(ErrReadOnly, f.path)
		}
	default:
		panic("unknown mode")
	}

	return nil
}

func (f *File) Name() string {
	return f.path
}

func (f *File) Write(p []byte) (int, error) {
	err := f.checkFlags(write)
	if err != nil {
		return 0, err
	}

	return f.g.Write(p)
}

func (f *File) Read(p []byte) (int, error) {
	err := f.checkFlags(read)
	if err != nil {
		return 0, err
	}

	n, err := f.g.Read(p)
	// on error n is negative
	if n < 0 {
		n = 0
	}
	// it does not tell when the file ended, if we could not read the whole
	// buffer treat it as EOF
	if err == nil && n < len(p) {
		err = io.EOF
	}

	return n, err
}

func (f *File) ReadAt(p []byte, off int64) (int, error) {
	err := f.checkFlags(read)
	if err != nil {
		return 0, err
	}

	offset, err := f.Seek(0, os.SEEK_CUR)
	if err != nil {
		return 0, err
	}

	n, err := f.g.ReadAt(p, off)
	if err == nil {
		_, err = f.Seek(offset, os.SEEK_SET)
		if err != nil {
			return 0, err
		}
	}

	// fix negative read bytes number and add EOF, same as Read
	if n < 0 {
		n = 0
	}
	if err == nil && n < len(p) {
		err = io.EOF
	}

	return n, err
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	return f.g.Seek(offset, whence)
}

func (f *File) Close() error {
	return f.g.Close()
}

func (f *File) Lock() error {
	return nil
}

func (f *File) Unlock() error {
	return nil
}

func (f *File) Truncate(size int64) error {
	return f.g.Truncate(size)
}
