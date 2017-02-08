package main

import (
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/ubuntu/display-snap/config"
)

const defaultimg = "nopreview.svg"

// VirtImages maps to real image file on disk if any. Default to nopreview.sgv
// It's compatible with http.FileServer()
type VirtImages struct{}

// Open map to real image file on disk if any. Default to nopreview.sgv
func (f VirtImages) Open(name string) (http.File, error) {
	var file http.File
	var err error
	for _, p := range []string{
		name,
		path.Join(config.Datadir, name),
		path.Join(config.Rootdir, name),
		path.Join(config.Rootdir, "www", defaultimg),
	} {
		if file, err = os.Open(p); err == nil {
			return file, nil
		}
	}

	return nil, fmt.Errorf("No image found")
}