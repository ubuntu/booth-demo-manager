package config

import (
	"log"
	"os"
	"path/filepath"
)

var (
	// Rootdir executable code to reach assets
	Rootdir string
	// Datadir access to write storage path
	Datadir string
)

func init() {
	// Set main set of directories
	var err error
	Rootdir = os.Getenv("SNAP")
	if Rootdir == "" {
		if Rootdir, err = filepath.Abs("."); err != nil {
			log.Fatal(err)
		}
	}
	Datadir = os.Getenv("SNAP_DATA")
	if Datadir == "" {
		Datadir = Rootdir
	}

}
