package datafs

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"strings"

	"github.com/hairyhenderson/go-fsimpl"
	"github.com/hairyhenderson/gomplate/v3/internal/config"
)

// // readMerge demultiplexes a `merge:` datasource. The 'args' parameter currently
// // has no meaning for this source.
// //
// // URI format is 'merge:<source 1>|<source 2>[|<source n>...]' where `<source #>`
// // is a supported URI or a pre-defined alias name.
// //
// // Query strings and fragments are interpreted relative to the merged data, not
// // the source data. To merge datasources with query strings or fragments, define
// // separate sources first and specify the alias names. HTTP headers are also not
// // supported directly.
// func (d *Data) readMerge(ctx context.Context, source *Source, args ...string) ([]byte, error) {
// 	opaque := source.URL.Opaque
// 	parts := strings.Split(opaque, "|")
// 	if len(parts) < 2 {
// 		return nil, fmt.Errorf("need at least 2 datasources to merge")
// 	}
// 	data := make([]map[string]interface{}, len(parts))
// 	for i, part := range parts {
// 		// supports either URIs or aliases
// 		subSource, err := d.lookupSource(part)
// 		if err != nil {
// 			// maybe it's a relative filename?
// 			u, uerr := config.ParseSourceURL(part)
// 			if uerr != nil {
// 				return nil, uerr
// 			}
// 			subSource = &Source{
// 				Alias: part,
// 				URL:   u,
// 			}
// 		}
// 		subSource.inherit(source)

// 		u := *subSource.URL

// 		base := path.Base(u.Path)
// 		if base == "/" {
// 			base = "."
// 		}

// 		u.Path = path.Dir(u.Path)

// 		fsp := datafs.FSProviderFromContext(ctx)

// 		fsys, err := fsp.New(&u)
// 		if err != nil {
// 			return nil, fmt.Errorf("lookup %s: %w", u.String(), err)
// 		}

// 		b, err := fs.ReadFile(fsys, base)
// 		if err != nil {
// 			return nil, fmt.Errorf("readFile (fs: %q, name: %q): %w", &u, base, err)
// 		}

// 		fi, err := fs.Stat(fsys, base)
// 		if err != nil {
// 			return nil, fmt.Errorf("stat (fs: %q, name: %q): %w", &u, base, err)
// 		}

// 		mimeType := fsimpl.ContentType(fi)

// 		data[i], err = parseMap(mimeType, string(b))
// 		if err != nil {
// 			return nil, err
// 		}
// 	}

// 	// Merge the data together
// 	b, err := mergeData(data)
// 	if err != nil {
// 		return nil, err
// 	}

// 	source.mediaType = yamlMimetype
// 	return b, nil
// }

// func mergeData(data []map[string]interface{}) (out []byte, err error) {
// 	dst := data[0]
// 	data = data[1:]

// 	dst, err = coll.Merge(dst, data...)
// 	if err != nil {
// 		return nil, err
// 	}

// 	s, err := ToYAML(dst)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return []byte(s), nil
// }

// func parseMap(mimeType, data string) (map[string]interface{}, error) {
// 	datum, err := parseData(mimeType, data)
// 	if err != nil {
// 		return nil, err
// 	}
// 	var m map[string]interface{}
// 	switch datum := datum.(type) {
// 	case map[string]interface{}:
// 		m = datum
// 	default:
// 		return nil, errors.Errorf("unexpected data type '%T' for datasource (type %s); merge: can only merge maps", datum, mimeType)
// 	}
// 	return m, nil
// }

// NewMergeFS returns a new filesystem that merges the contents of multiple
// paths. Only a URL like "merge:" or "merge:///" makes sense here - the
// piped-separated lists of sub-sources to merge must be given to Open.
//
// Usually you'll want to use WithDataSourcesFS to provide the map of
// datasources that can be referenced. Otherwise, only URLs will be supported.
//
// An FSProvider will also be needed, which can be provided with a context
// using ContextWithFSProvider. Provide that context with fsimpl.WithContextFS.
func NewMergeFS(u *url.URL) (fs.FS, error) {
	if u.Scheme != "merge" {
		return nil, fmt.Errorf("unsupported scheme %q", u.Scheme)
	}

	return &mergeFS{
		ctx:     context.Background(),
		sources: map[string]config.DataSource{},
	}, nil
}

type mergeFS struct {
	ctx        context.Context
	httpClient *http.Client
	sources    map[string]config.DataSource
}

//nolint:gochecknoglobals
var MergeFS = fsimpl.FSProviderFunc(NewMergeFS, "merge")

var (
	_ fs.FS             = (*mergeFS)(nil)
	_ withContexter     = (*mergeFS)(nil)
	_ withDataSourceser = (*mergeFS)(nil)
)

func (f mergeFS) WithContext(ctx context.Context) fs.FS {
	fsys := f
	fsys.ctx = ctx

	return &fsys
}

func (f mergeFS) WithHTTPClient(client *http.Client) fs.FS {
	fsys := f
	fsys.httpClient = client

	return &fsys
}

func (f mergeFS) WithDataSources(sources map[string]config.DataSource) fs.FS {
	fsys := f
	fsys.sources = sources

	return &fsys
}

func (f *mergeFS) Open(name string) (fs.File, error) {
	parts := strings.Split(name, "|")
	if len(parts) < 2 {
		return nil, &fs.PathError{Op: "open", Path: name,
			Err: fmt.Errorf("need at least 2 datasources to merge"),
		}
	}

	fsp := FSProviderFromContext(f.ctx)

	// now open each of the sub-files
	subFiles := make([]fs.File, len(parts))

	for i, part := range parts {
		// if this is a datasource, look it up
		subSource, ok := f.sources[part]
		if !ok {
			// maybe it's a relative filename?
			u, uerr := config.ParseSourceURL(part)
			if uerr != nil {
				return nil, uerr
			}
			subSource = config.DataSource{URL: u}
		}

		fsURL, base := SplitFSMuxURL(subSource.URL)

		fsys, err := fsp.New(fsURL)
		if err != nil {
			return nil, &fs.PathError{Op: "open", Path: name,
				Err: fmt.Errorf("lookup for %s: %w", subSource.URL.String(), err),
			}
		}

		// pass in the context and other bits
		fsys = fsimpl.WithContextFS(f.ctx, fsys)
		fsys = fsimpl.WithHeaderFS(subSource.Header, fsys)
		fsys = fsimpl.WithHTTPClientFS(f.httpClient, fsys)
		fsys = WithDataSourcesFS(f.sources, fsys)

		// now open the file
		f, err := fsys.Open(base)
		if err != nil {
			return nil, &fs.PathError{Op: "open", Path: name,
				Err: fmt.Errorf("opening merge part %q: %w", part, err),
			}
		}

		subFiles[i] = f
	}

	return &mergeFile{name: name, subFiles: subFiles}, nil
}

type mergeFile struct {
	name     string
	subFiles []fs.File
}

var _ fs.File = (*mergeFile)(nil)

func (f *mergeFile) Close() error {
	for _, f := range f.subFiles {
		f.Close()
	}
	return nil
}

func (f *mergeFile) Stat() (fs.FileInfo, error) {
	return nil, fmt.Errorf("not implemented")
}

func (f *mergeFile) Read(p []byte) (int, error) {
	// read from all and merge here..

	return 0, fmt.Errorf("not implemented")
}

type fileContent struct {
	contentType string
	b           []byte
}

// not quite there yet...
func mergeData(data []map[string]interface{}) (out []byte, err error) {
	return nil, fmt.Errorf("mergeData: not implemented")
}
