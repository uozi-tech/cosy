package app

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed dist/*
var distFS embed.FS

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
	subFS, _ := fs.Sub(distFS, "dist")
	return &EmbedFileSystem{
		FileSystem: http.FS(subFS),
		subFS:     subFS,
	}
}

// HTTPFileSystem returns an http.FileSystem that serves from the embedded filesystem (for backward compatibility)
func HTTPFileSystem() http.FileSystem {
	subFS, _ := fs.Sub(distFS, "dist")
	return http.FS(subFS)
}

// Open opens a file from the embedded filesystem
func Open(name string) (fs.File, error) {
	name = strings.TrimPrefix(name, "/")
	subFS, _ := fs.Sub(distFS, "dist")
	return subFS.Open(name)
}
