package cmd

import (
	"context"

	"github.com/hairyhenderson/go-fsimpl"
	"github.com/hairyhenderson/go-fsimpl/awssmfs"
	"github.com/hairyhenderson/go-fsimpl/awssmpfs"
	"github.com/hairyhenderson/go-fsimpl/blobfs"
	"github.com/hairyhenderson/go-fsimpl/consulfs"
	"github.com/hairyhenderson/go-fsimpl/filefs"
	"github.com/hairyhenderson/go-fsimpl/gitfs"
	"github.com/hairyhenderson/go-fsimpl/httpfs"
	"github.com/hairyhenderson/go-fsimpl/vaultfs"
	"github.com/hairyhenderson/gomplate/v3/internal/datafs"
)

func injectFSProviders(ctx context.Context) context.Context {
	if datafs.FSProviderFromContext(ctx) != nil {
		return ctx
	}

	// inject a go-fsimpl filesystem provider if it hasn't already been
	// overridden
	fsp := fsimpl.NewMux()
	// go-fsimpl filesystems (same as autofs)
	fsp.Add(awssmfs.FS)
	fsp.Add(awssmpfs.FS)
	fsp.Add(blobfs.FS)
	fsp.Add(consulfs.FS)
	fsp.Add(filefs.FS)
	fsp.Add(gitfs.FS)
	fsp.Add(httpfs.FS)
	fsp.Add(vaultfs.FS)
	// custom gomplate filesystem(s)
	fsp.Add(datafs.EnvFS)
	fsp.Add(datafs.StdinFS)
	fsp.Add(datafs.MergeFS)

	return datafs.ContextWithFSProvider(ctx, fsp)
}
