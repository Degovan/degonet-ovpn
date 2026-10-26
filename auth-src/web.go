package main

import (
	"embed"
	"io/fs"
)

// Hasil `npm run build` di web/. Folder ini selalu berisi .gitkeep supaya `go build`
// tetap jalan walau frontend belum di-build.
//
//go:embed all:web/dist
var webDist embed.FS

func webFS() fs.FS {
	sub, err := fs.Sub(webDist, "web/dist")
	if err != nil {
		panic(err)
	}
	return sub
}
