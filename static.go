package static // import "github.com/novakit/static"

import (
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/novakit/binfs"
	"github.com/novakit/nova"
)

// Options options for static
type Options struct {
	// Prefix path prefix of url
	Prefix string
	// Directory directory to serve, will use WEBROOT envvar
	Directory string
	// BinFS using binfs, not suggested with WEBROOT envvar
	BinFS bool
}

// ResponseWriterWrapper wrapper of http.ResponseWriter, blocks Write() if 404 or 403
type ResponseWriterWrapper struct {
	http.ResponseWriter
	// TempHeader temporary header
	TempHeader http.Header
	// Blocked if wrapper is blocked
	Blocked bool
	// HeaderWritten if header is already written
	HeaderWritter bool
}

// Header returns the temporary header
func (r *ResponseWriterWrapper) Header() http.Header {
	return r.TempHeader
}

// WriteHeader override http.ResponseWriter
func (r *ResponseWriterWrapper) WriteHeader(statusCode int) {
	if r.HeaderWritter {
		return
	}
	r.HeaderWritter = true

	if statusCode == http.StatusForbidden || statusCode == http.StatusNotFound {
		r.Blocked = true
		return
	}
	// write back Header
	for k, v := range r.TempHeader {
		r.ResponseWriter.Header()[k] = v
	}
	// invoke original response writer
	r.ResponseWriter.WriteHeader(statusCode)
}

// Write override http.ResponseWriter
func (r *ResponseWriterWrapper) Write(p []byte) (int, error) {
	if !r.HeaderWritter {
		r.WriteHeader(http.StatusOK)
	}
	if r.Blocked {
		return len(p), nil
	}
	return r.ResponseWriter.Write(p)
}

func sanitizeOptions(opts ...Options) (opt Options) {
	if len(opts) > 0 {
		opt = opts[0]
	}
	if len(opt.Directory) == 0 {
		opt.Directory = os.Getenv("WEBROOT")
	}
	return
}

func trimPathPrefix(pfx string, path string) (ret string, ok bool) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if !strings.HasPrefix(pfx, "/") {
		pfx = "/" + pfx
	}
	if strings.HasPrefix(path, pfx) {
		ret = "/" + path[len(pfx):]
		ok = true
		return
	}
	ret = path
	return
}

func cloneRequest(r *http.Request) *http.Request {
	r2 := new(http.Request)
	*r2 = *r
	// Deep copy the URL because it isn't
	// a map and the URL is mutable by users
	// of WithContext.
	if r.URL != nil {
		r2URL := new(url.URL)
		*r2URL = *r.URL
		r2.URL = r2URL
	}
	return r2
}

func buildFileServer(opt Options) http.Handler {
	var fs http.Handler
	if opt.BinFS {
		c := strings.Split(opt.Directory, "/")
		n := binfs.Find(c...)
		if n == nil {
			panic("directory not find in binfs")
		}
		fs = http.FileServer(n.FileSystem())
	} else {
		fs = http.FileServer(http.Dir(opt.Directory))
	}
	return fs
}

// Handler create a nova.HandlerFunc
func Handler(opts ...Options) nova.HandlerFunc {
	opt := sanitizeOptions(opts...)
	fs := buildFileServer(opt)
	return func(c *nova.Context) (err error) {
		// must be GET/HEAD method
		if c.Req.Method != http.MethodGet && c.Req.Method != http.MethodHead {
			c.Next()
			return
		}
		// validate and trim prefix
		var req = c.Req
		if len(opt.Prefix) > 0 {
			if p, ok := trimPathPrefix(opt.Prefix, req.URL.Path); ok {
				// clone request and update URL.Path
				req = cloneRequest(req)
				req.URL.Path = p
			} else {
				// skip if prefix mismatch
				c.Next()
				return
			}
		}
		// invoke http.FileServer with a wrapped http.ResponseWriter
		res := &ResponseWriterWrapper{ResponseWriter: c.Res, TempHeader: http.Header{}}
		fs.ServeHTTP(res, req)
		// if blocked (404 or 403), c.Res is untouched, invoke next handler
		if res.Blocked {
			c.Next()
		}
		return
	}
}
