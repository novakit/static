package static_test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/novakit/nova"
	"github.com/novakit/static"
	"github.com/novakit/testkit"
)

func TestStatic_BinFS(t *testing.T) {
	n := nova.New()
	n.Use(static.Handler(static.Options{
		Prefix:    "static",
		Directory: "testdata",
		BinFS:     true,
	}))
	n.Use(func(c *nova.Context) error {
		c.Res.Write([]byte("NOT FOUND" + c.Req.URL.Path))
		return nil
	})
	req, _ := http.NewRequest(http.MethodGet, "/static/dir2/dir21/file212.js", nil)
	res := testkit.NewDummyResponse()
	n.ServeHTTP(res, req)
	// should serve file
	if !bytes.Equal(res.Bytes(), binfs0e2b285092f29e6844cf004e91ee596a4f392d82.Data) {
		t.Error("request failed 1", res.String())
	}
	// should pass through without modifying request.URL
	req, _ = http.NewRequest(http.MethodGet, "/static/dir2/dir21/file212.notexist.js", nil)
	res = testkit.NewDummyResponse()
	n.ServeHTTP(res, req)
	if res.String() != "NOT FOUND/static/dir2/dir21/file212.notexist.js" {
		t.Error("request failed 2", res.String())
	}
}

func TestStatic_Dir(t *testing.T) {
	n := nova.New()
	n.Use(static.Handler(static.Options{
		Prefix:    "static",
		Directory: "testdata",
	}))
	n.Use(func(c *nova.Context) error {
		c.Res.Write([]byte("NOT FOUND" + c.Req.URL.Path))
		return nil
	})
	req, _ := http.NewRequest(http.MethodGet, "/static/dir2/dir21/file212.js", nil)
	res := testkit.NewDummyResponse()
	n.ServeHTTP(res, req)
	// should serve file
	if !bytes.Equal(res.Bytes(), binfs0e2b285092f29e6844cf004e91ee596a4f392d82.Data) {
		t.Error("request failed 1", res.String())
	}
	// should fallback to next handler
	req, _ = http.NewRequest(http.MethodGet, "/static/dir2/dir21/file212.notexist.js", nil)
	res = testkit.NewDummyResponse()
	n.ServeHTTP(res, req)
	// should pass through without modifying request.URL
	if res.String() != "NOT FOUND/static/dir2/dir21/file212.notexist.js" {
		t.Error("request failed 2", res.String())
	}
}

func TestStatic_DirWithIndex(t *testing.T) {
	n := nova.New()
	n.Use(static.Handler(static.Options{
		Prefix:    "static",
		Directory: "testdata",
		Index:     true,
	}))
	n.Use(func(c *nova.Context) error {
		c.Res.Write([]byte("NOT FOUND" + c.Req.URL.Path))
		return nil
	})
	req, _ := http.NewRequest(http.MethodGet, "/static/dir1", nil)
	res := testkit.NewDummyResponse()
	n.ServeHTTP(res, req)
	// should serve file
	if !bytes.Equal(res.Bytes(), binfs4f9a993c2ae5cfcfd698da51d05bc646a2ed571b.Data) {
		t.Error("request failed 1", res.String())
	}
}
