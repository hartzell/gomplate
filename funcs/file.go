package funcs

import (
	"context"
	"io/fs"
	"os"

	"github.com/hack-pad/hackpadfs"
	"github.com/hairyhenderson/gomplate/v3/conv"
	"github.com/hairyhenderson/gomplate/v3/file"
)

// FileNS - the File namespace
//
// Deprecated: don't use
func FileNS() *FileFuncs {
	return &FileFuncs{}
}

// AddFileFuncs -
//
// Deprecated: use [CreateFileFuncs] instead
func AddFileFuncs(f map[string]interface{}) {
	for k, v := range CreateFileFuncs(context.Background()) {
		f[k] = v
	}
}

// CreateFileFuncs -
func CreateFileFuncs(ctx context.Context) map[string]interface{} {
	ns := &FileFuncs{
		ctx: ctx,
		fs:  datafs.WrapWdFS(hackpadfs.NewOsFS())),
	}
	return map[string]interface{}{
		"file": func() interface{} { return ns },
	}
}

// FileFuncs -
type FileFuncs struct {
	ctx context.Context
	fs  afero.Fs
}

// Read -
func (f *FileFuncs) Read(path interface{}) (string, error) {
	return file.Read(conv.ToString(path))
}

// Stat -
func (f *FileFuncs) Stat(path interface{}) (fs.FileInfo, error) {
	return f.fs.Stat(conv.ToString(path))
}

// Exists -
func (f *FileFuncs) Exists(path interface{}) bool {
	_, err := f.Stat(conv.ToString(path))
	return err == nil
}

// IsDir -
func (f *FileFuncs) IsDir(path interface{}) bool {
	i, err := f.Stat(conv.ToString(path))
	return err == nil && i.IsDir()
}

// ReadDir -
func (f *FileFuncs) ReadDir(path interface{}) ([]string, error) {
	return file.ReadDir(conv.ToString(path))
}

// Walk -
func (f *FileFuncs) Walk(path interface{}) ([]string, error) {
	files := make([]string, 0)
	err := afero.Walk(f.fs, conv.ToString(path), func(subpath string, finfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		files = append(files, subpath)
		return nil
	})
	return files, err
}

// Write -
func (f *FileFuncs) Write(path interface{}, data interface{}) (s string, err error) {
	if b, ok := data.([]byte); ok {
		err = file.Write(conv.ToString(path), b)
	} else {
		err = file.Write(conv.ToString(path), []byte(conv.ToString(data)))
	}
	return "", err
}
