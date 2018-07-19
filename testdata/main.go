package main

import (
	"net/http"

	"github.com/novakit/nova"
	"github.com/novakit/static"
)

func main() {
	n := nova.New()
	n.Use(static.Handler(static.Options{
		Directory: ".",
	}))
	http.ListenAndServe(":9999", n)
}
