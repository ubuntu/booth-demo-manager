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

package pilot

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path"

	"github.com/ubuntu/booth-demo-manager/config"

	yaml "gopkg.in/yaml.v2"
)

// CurrentDemoMsg element with ID, its current URL and optional index
type CurrentDemoMsg struct {
	ID    string
	Index int
	URL   string
	Auto  bool
}

var (
	allDemos     map[string]Demo
	current      *CurrentDemo
	demoFilePath *string
)

const (
	demoDefaultFilename = "demos.def"
	defaultTime         = 30
)

func init() {
	demoFilePath = flag.String("c", demoDefaultFilename, "config file path overriding default one")
}

// Start all demos. Return a channel of current demo ID
// and all demos
// TODO: starts and close it properly once we can shutdown webserver
func Start(changeCurrent <-chan CurrentDemoMsg) (<-chan CurrentDemoMsg, <-chan map[string]Demo, error) {
	currentCh := make(chan CurrentDemoMsg)
	allDemosCh := make(chan map[string]Demo)

	if err := loadDefinition(); err != nil {
		return nil, nil, err
	}

	go func() {
		// sending first all Demos list
		allDemosCh <- allDemos

		for {
			select {
			case elem := <-changeCurrent:
				d, ok := allDemos[elem.ID]
				if !ok {
					log.Printf("%s not in currently available demos", elem.ID)
					continue
				}
				// We avoid a potential race, waiting for the older current object to be deselected before selecting the new one
				// Especially important when the same one is selected again
				if current != nil {
					current.Release()
					<-current.deselected
				}
				current = d.Select(elem.ID, elem.Index, currentCh)
			}
		}

	}()

	return currentCh, allDemosCh, nil
}

func sendNewCurrentURL(ch chan<- CurrentDemoMsg, c *CurrentDemo) {
	ch <- CurrentDemoMsg{ID: c.id, URL: c.url, Index: c.slideIndex, Auto: c.auto}
}

func loadDefinition() error {
	var data []byte
	var err error
	var selectedFile string

	// Always look for relative path first.
	potentialDemoFiles := []string{*demoFilePath}
	// If default name, look for more places.
	if *demoFilePath == demoDefaultFilename {
		potentialDemoFiles = append(potentialDemoFiles,
			path.Join(config.Datadir, demoDefaultFilename),
			path.Join(config.Rootdir, demoDefaultFilename))
	}

	for _, selectedFile := range potentialDemoFiles {
		data, err = ioutil.ReadFile(selectedFile)
		if err == nil {
			break
		}
	}
	if data == nil {
		return fmt.Errorf("Couldn't read any of config file as: %v", potentialDemoFiles)
	}

	allDemos = make(map[string]Demo)
	if err := yaml.Unmarshal(data, &allDemos); err != nil {
		return fmt.Errorf("%s isn't a valid yaml file: %v", selectedFile, err)
	}

	// remove invalid elements and set default timer
	for id, d := range allDemos {
		if d.URL == "" && len(d.Slides) == 0 {
			fmt.Printf("Removing %s has no url nor slides attributes\n", id)
			delete(allDemos, id)
		}
		if len(d.Slides) > 0 && d.Time == 0 {
			d.Time = defaultTime
			allDemos[id] = d
		}
		if d.URL != "" && len(d.Slides) > 0 {
			fmt.Printf("%s has both url and slides attributes. Will only use slides\n", id)
		}
	}

	return nil
}
