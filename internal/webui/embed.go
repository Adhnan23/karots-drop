package webui

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed index.html file.html retrieve.html docs.html static
var assets embed.FS

func Open(name string) (fs.File, error) {
	return assets.Open(name)
}

func FileSystem() http.FileSystem {
	return http.FS(assets)
}

func StaticFS() http.FileSystem {
	sub, err := fs.Sub(assets, "static")
	if err != nil {
		panic(err)
	}
	return http.FS(sub)
}
