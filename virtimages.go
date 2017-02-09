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

package main

import (
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/ubuntu/booth-demo-manager/config"
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
