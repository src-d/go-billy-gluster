package gluster

import (
	"os"
	"path/filepath"
	"syscall"

	"github.com/gluster/gogfapi/gfapi"
	"gopkg.in/src-d/go-billy.v4"
)

const (
	strNoFileOrDir = "no such file or directory"

	defaultDirectoryMode = 0755
	defaultCreateMode    = 0666

	capabilities billy.Capability = billy.WriteCapability |
		billy.ReadCapability | billy.ReadAndWriteCapability |
		billy.SeekCapability | billy.TruncateCapability
)

var _ billy.Basic = new(FS)

// FS manages the filesystem of a gluster volume. It implements
// billy.Basic.
type FS struct {
	v *gfapi.Volume
}

// New creates a new FS connecting it to the specified host and volume.
func New(host, volume string) (*FS, error) {
	vol := new(gfapi.Volume)
	err := vol.Init(volume, host)
	if err != nil {
		return nil, err
	}

	err = vol.Mount()
	if err != nil {
		return nil, err
	}

	g := &FS{v: vol}
	return g, nil
}

// Close unmounts the gluster volume associated to the FS.
func (g *FS) Close() error {
	return g.v.Unmount()
}

// Create implements billy.Basic interface.
func (g *FS) Create(filename string) (billy.File, error) {
	if err := g.createDir(filename); err != nil {
		return nil, err
	}

	f, err := g.v.Create(filename)
	if err != nil {
		return nil, err
	}

	return NewFile(filename, f, os.O_RDWR), nil
}

// Open implements billy.Basic interface.
func (g *FS) Open(filename string) (billy.File, error) {
	return g.OpenFile(filename, os.O_RDONLY, 0)
}

// OpenFile implements billy.Basic interface.
func (g *FS) OpenFile(
	filename string,
	flag int,
	perm os.FileMode,
) (billy.File, error) {
	if flag&os.O_CREATE == os.O_CREATE {
		if err := g.createDir(filename); err != nil {
			return nil, err
		}

		// O_CREATE does not create the file. Here Create is used if we can
		// not find it. This could be done in a more efficient way by reusing
		// the created file descriptor instead of reopening with the specific
		// flags in some cases.
		_, err := g.Stat(filename)
		if err != nil {
			c, err := g.v.Create(filename)
			if err != nil {
				return nil, err
			}
			if err = c.Close(); err != nil {
				return nil, err
			}

			// Setting permissions in OpenFile is not supported. Change it
			// manually with Chmod.
			err = g.v.Chmod(filename, perm)
			if err != nil {
				return nil, err
			}
		}
	}

	f, err := g.v.OpenFile(filename, flag, perm)
	if err != nil {
		return nil, err
	}

	return NewFile(filename, f, flag), nil
}

// Stat implements billy.Basic interface.
func (g *FS) Stat(filename string) (os.FileInfo, error) {
	return g.v.Stat(filename)
}

// Rename implements billy.Basic interface.
func (g *FS) Rename(oldpath string, newpath string) error {
	if err := g.createDir(newpath); err != nil {
		return err
	}

	return g.v.Rename(oldpath, newpath)
}

// Remove implements billy.Basic interface.
func (g *FS) Remove(filename string) error {
	err := g.v.Unlink(filename)
	if err == nil {
		return nil
	}

	if e, ok := err.(*os.PathError); ok {
		if e.Err == syscall.EISDIR {
			return g.v.Rmdir(filename)
		}
	}

	return err
}

// Join implements billy.Basic interface.
func (g *FS) Join(elem ...string) string {
	return filepath.Join(elem...)
}

// ReadDir implements billy.Dir interface.
func (g *FS) ReadDir(path string) ([]os.FileInfo, error) {
	d, err := g.v.Open(path)
	if err != nil {
		return nil, err
	}

	defer d.Close()

	files, err := d.Readdir(0)
	if err != nil {
		return nil, err
	}

	// gluster Readdir returns also "." and ".."
	n := len(files)
	for i := 0; i < n; i++ {
		if files[i].Name() == "." || files[i].Name() == ".." {
			// swap with the last element
			files[i], files[n-1] = files[n-1], files[i]
			n--
			i--
		}
	}

	return files[:n], nil
}

// MkdirAll implements billy.Dir interface.
func (g *FS) MkdirAll(filename string, perm os.FileMode) error {
	return g.v.MkdirAll(filename, perm)
}

// Capabilities implements billy.Capable interface.
func (g *FS) Capabilities() billy.Capability {
	return capabilities
}

func (g *FS) createDir(fullpath string) error {
	dir := filepath.Dir(fullpath)
	if dir != "." {
		if err := g.MkdirAll(dir, defaultDirectoryMode); err != nil {
			return err
		}
	}

	return nil
}
