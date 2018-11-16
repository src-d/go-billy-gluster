package gluster

import (
	"os"
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

	FS  *GlusterFS
	tmp string
}

func (s *FilesystemSuite) SetUpTest(c *C) {
	fs, err := New("172.17.0.2", "billy")
	c.Assert(err, IsNil)

	s.tmp, err = util.TempDir(fs, "", "billy")
	c.Assert(err, IsNil)

	tmp := chroot.New(fs, s.tmp)
	s.BasicSuite.FS = tmp
}

func (s *FilesystemSuite) TearDownTest(c *C) {
	if s.FS != nil {
		err := os.RemoveAll(s.tmp)
		c.Assert(err, IsNil)

		err = s.FS.Close()
		c.Assert(err, IsNil)
	}
}
