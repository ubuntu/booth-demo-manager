/*
Copyright 2017 Canonical Ltd.
This file is part of booth-demo-manager.

booth-demo-manager is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, version 3 of the License.

Foobar is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with booth-demo-manager.  If not, see <http://www.gnu.org/licenses/>.
*/

package config

import (
	"log"
	"os"
	"path"
	"path/filepath"
)

var (
	// Rootdir executable code to reach assets
	Rootdir string
	// Datadir access to write storage path
	Datadir string
	// DemoBaseDir access to potential demo directory
	DemoBaseDir string
)

func init() {
	// Set main set of directories
	var err error
	Rootdir = os.Getenv("SNAP")
	if Rootdir == "" {
		if Rootdir, err = filepath.Abs("."); err != nil {
			log.Fatal(err)
		}
	} else {
		DemoBaseDir = path.Join(Rootdir, "..", "..")
	}
	Datadir = os.Getenv("SNAP_DATA")
	if Datadir == "" {
		Datadir = Rootdir
	}

}
