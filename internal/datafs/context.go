package datafs

import (
	"context"
	"io"
	"io/fs"
	"os"

	"github.com/hairyhenderson/go-fsimpl"
	"github.com/hairyhenderson/gomplate/v3/internal/config"
)

type fsProviderCtxKey struct{}

// ContextWithFSProvider returns a context with the given [fsimpl.FSProvider]
func ContextWithFSProvider(ctx context.Context, fsp fsimpl.FSProvider) context.Context {
	return context.WithValue(ctx, fsProviderCtxKey{}, fsp)
}

// FSProviderFromContext returns the FSProvider from the context, if any
func FSProviderFromContext(ctx context.Context) fsimpl.FSProvider {
	if fsp, ok := ctx.Value(fsProviderCtxKey{}).(fsimpl.FSProvider); ok {
		return fsp
	}

	return nil
}

// withContexter is an fs.FS that can be configured with a custom context
// copied from go-fsimpl - see internal/types.go
type withContexter interface {
	WithContext(ctx context.Context) fs.FS
}

type withDataSourceser interface {
	WithDataSources(sources map[string]config.DataSource) fs.FS
}

// WithDataSourcesFS injects a datasource map into the filesystem fs, if the
// filesystem supports it (i.e. has a WithDataSources method). This is used for
// the mergefs filesystem.
func WithDataSourcesFS(sources map[string]config.DataSource, fsys fs.FS) fs.FS {
	if fsys, ok := fsys.(withDataSourceser); ok {
		return fsys.WithDataSources(sources)
	}

	return fsys
}

type stdinCtxKey struct{}

func ContextWithStdin(ctx context.Context, r io.Reader) context.Context {
	return context.WithValue(ctx, stdinCtxKey{}, r)
}

func StdinFromContext(ctx context.Context) io.Reader {
	if r, ok := ctx.Value(stdinCtxKey{}).(io.Reader); ok {
		return r
	}

	return os.Stdin
}
