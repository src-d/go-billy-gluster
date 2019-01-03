package gluster

import (
	"testing"

	. "gopkg.in/check.v1"
	"gopkg.in/src-d/go-billy.v4/helper/chroot"
	"gopkg.in/src-d/go-billy.v4/test"
	"gopkg.in/src-d/go-billy.v4/util"
)

func Test(t *testing.T) { TestingT(t) }

var _ = Suite(&FilesystemSuite{})

type FilesystemSuite struct {
	test.BasicSuite
	test.DirSuite

	FS  *FS
	tmp string
}

func (s *FilesystemSuite) SetUpTest(c *C) {
	fs, err := New("localhost", "billy")
	c.Assert(err, IsNil)
	s.FS = fs

	s.tmp, err = util.TempDir(fs, "", "billy")
	c.Assert(err, IsNil)

	tmp := chroot.New(fs, s.tmp)
	s.BasicSuite.FS = tmp
	s.DirSuite.FS = tmp
}

func (s *FilesystemSuite) TearDownTest(c *C) {
	if s.FS != nil {
		err := util.RemoveAll(s.FS, s.tmp)
		c.Assert(err, IsNil)

		err = s.FS.Close()
		c.Assert(err, IsNil)
	}
}

func (s *FilesystemSuite) TestReaddirEmpty(c *C) {
	fs := s.DirSuite.FS
	err := fs.MkdirAll("test", 0777)
	c.Assert(err, IsNil)

	files, err := fs.ReadDir("test")
	c.Assert(err, IsNil)
	c.Assert(len(files), Equals, 0)
}
