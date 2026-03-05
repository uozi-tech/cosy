package app

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed index.html css js fonts assets
var staticFS embed.FS

// EmbedFileSystem implements static.ServeFileSystem interface
type EmbedFileSystem struct {
	http.FileSystem
	subFS fs.FS
}

// Exists checks if a file exists at the given path
func (efs *EmbedFileSystem) Exists(prefix, path string) bool {
	path = strings.TrimPrefix(path, "/")
	if after, found := strings.CutPrefix(path, prefix); found {
		path = strings.TrimPrefix(after, "/")
	}

	_, err := efs.subFS.Open(path)
	return err == nil
}

// ServeFileSystem returns a static.ServeFileSystem that serves from the embedded filesystem
func ServeFileSystem() *EmbedFileSystem {
	return &EmbedFileSystem{
		FileSystem: http.FS(staticFS),
		subFS:      staticFS,
	}
}

// HTTPFileSystem returns an http.FileSystem that serves from the embedded filesystem
func HTTPFileSystem() http.FileSystem {
	return http.FS(staticFS)
}

// Open opens a file from the embedded filesystem
func Open(name string) (fs.File, error) {
	name = strings.TrimPrefix(name, "/")
	return staticFS.Open(name)
}
