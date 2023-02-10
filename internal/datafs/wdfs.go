package datafs

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/hack-pad/hackpadfs"
)

// wdFS is a filesystem wrapper that assumes non-absolute paths are relative to
// the current working directory (at time of wrapping).
// It only works in a meaningful way when used  with a local filesystem (e.g.
// [os.DirFS] or [hackpadfs/os.FS]).
// If os.Chdir() is called, the wrapped filesystem will still be relative to the
// original working directory.
type wdFS struct {
	fsys fs.FS
	cwd  string
}

var (
	_ fs.FS                = &wdFS{}
	_ fs.StatFS            = &wdFS{}
	_ fs.ReadFileFS        = &wdFS{}
	_ fs.ReadDirFS         = &wdFS{}
	_ fs.SubFS             = &wdFS{}
	_ fs.GlobFS            = &wdFS{}
	_ hackpadfs.CreateFS   = &wdFS{}
	_ hackpadfs.OpenFileFS = &wdFS{}
	_ hackpadfs.MkdirFS    = &wdFS{}
	_ hackpadfs.MkdirAllFS = &wdFS{}
	_ hackpadfs.RemoveFS   = &wdFS{}
)

func WrapWdFS(fsys fs.FS) (fs.FS, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("getwd: %w", err)
	}

	return &wdFS{fsys: fsys, cwd: cwd}, nil
}

func (w *wdFS) resolve(name string) string {
	if !filepath.IsAbs(name) {
		name = filepath.Join(w.cwd, name)
	}

	return name[1:]
}

func (w *wdFS) Open(name string) (fs.File, error) {
	return w.fsys.Open(w.resolve(name))
}

func (w *wdFS) Stat(name string) (fs.FileInfo, error) {
	return fs.Stat(w.fsys, w.resolve(name))
}

func (w *wdFS) ReadFile(name string) ([]byte, error) {
	return fs.ReadFile(w.fsys, w.resolve(name))
}

func (w *wdFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return fs.ReadDir(w.fsys, w.resolve(name))
}

func (w *wdFS) Sub(name string) (fs.FS, error) {
	return fs.Sub(w.fsys, w.resolve(name))
}

func (w *wdFS) Glob(pattern string) ([]string, error) {
	return fs.Glob(w.fsys, w.resolve(pattern))
}

func (w *wdFS) Create(name string) (fs.File, error) {
	return hackpadfs.Create(w.fsys, w.resolve(name))
}

func (w *wdFS) OpenFile(name string, flag int, perm fs.FileMode) (fs.File, error) {
	return hackpadfs.OpenFile(w.fsys, w.resolve(name), flag, perm)
}

func (w *wdFS) Mkdir(name string, perm fs.FileMode) error {
	return hackpadfs.Mkdir(w.fsys, w.resolve(name), perm)
}

func (w *wdFS) MkdirAll(name string, perm fs.FileMode) error {
	return hackpadfs.MkdirAll(w.fsys, w.resolve(name), perm)
}

func (w *wdFS) Remove(name string) error {
	return hackpadfs.Remove(w.fsys, w.resolve(name))
}
