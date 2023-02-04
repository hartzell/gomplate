//go:build !windows

package gomplate

import (
	"context"
	"io/fs"
	"testing"

	"github.com/hack-pad/hackpadfs"
	"github.com/hack-pad/hackpadfs/mem"
	"github.com/hairyhenderson/gomplate/v3/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWalkDir(t *testing.T) {
	ctx := context.Background()

	var fsys fs.FS
	fsys, _ = mem.NewFS()
	fsys, _ = wrapWdFS(fsys)

	cfg := &config.Config{}

	_, err := walkDir(ctx, fsys, cfg, "/indir", simpleNamer("/outdir"), nil, 0, false)
	assert.Error(t, err)

	err = hackpadfs.MkdirAll(fsys, "/indir/one", 0777)
	require.NoError(t, err)
	err = hackpadfs.MkdirAll(fsys, "/indir/two", 0777)
	require.NoError(t, err)
	hackpadfs.WriteFullFile(fsys, "/indir/one/foo", []byte("foo"), 0o644)
	hackpadfs.WriteFullFile(fsys, "/indir/one/bar", []byte("bar"), 0o644)
	hackpadfs.WriteFullFile(fsys, "/indir/two/baz", []byte("baz"), 0o644)

	templates, err := walkDir(ctx, fsys, cfg, "/indir", simpleNamer("/outdir"), []string{"*/two"}, 0, false)
	require.NoError(t, err)

	expected := []Template{
		{
			Name: "/indir/one/bar",
			Text: "bar",
		},
		{
			Name: "/indir/one/foo",
			Text: "foo",
		},
	}
	assert.Len(t, templates, 2)
	for i, tmpl := range templates {
		assert.Equal(t, expected[i].Name, tmpl.Name)
		assert.Equal(t, expected[i].Text, tmpl.Text)
	}
}
