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
	"os"
	"path"
	"path/filepath"
	"strings"

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
	demoDefaultFilename = "booth-demo-manager.def"
	defaultTime         = 30
)

func init() {
	demoFilePath = flag.String("c", demoDefaultFilename, "config file path overriding default one and autodetection")
	allDemos = make(map[string]Demo)
}

// Start all demos. Return a channel of current demo ID
// and all demos
// TODO: starts and close it properly once we can shutdown webserver
func Start(changeCurrent <-chan CurrentDemoMsg, startPageURL string) (<-chan CurrentDemoMsg, <-chan map[string]Demo, error) {
	currentCh := make(chan CurrentDemoMsg)
	allDemosCh := make(chan map[string]Demo)

	if err := loadAllDemos(&allDemos, startPageURL); err != nil {
		return nil, nil, err
	}

	go func() {
		// sending first all Demos list and start page
		allDemosCh <- allDemos
		current = allDemos[""].Select("", -1, currentCh)

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

func loadAllDemos(allDemos *map[string]Demo, startPageURL string) error {
	var err error

	// Always look for relative path first.
	var configFiles []string

	// If specified on the command line, take only that file.
	if *demoFilePath != demoDefaultFilename {
		configFiles = append(configFiles, *demoFilePath)
	} else {
		// If default name, look for more places, including autodetection.
		// Last one wins over others.
		// Rootdir
		// SnapDir
		// Relative path
		// Autodetect config
		configFiles = append(configFiles,
			path.Join(config.Rootdir, demoDefaultFilename),
			path.Join(config.Datadir, demoDefaultFilename),
			demoDefaultFilename)

		// try to detect files for every installed demos
		if detectedConfigs, err := getValidDemosConfig(config.DemoBaseDir); err != nil {
			log.Printf("Auto demo config loading error: %v", err)
		} else {
			for _, c := range detectedConfigs {
				configFiles = append(configFiles, c)
			}
		}
	}

	// load all demos from config
	for _, configFile := range configFiles {
		if err = loadDemoAndSanitize(allDemos, configFile); err != nil {
			log.Println(err)
		}
	}

	// Add start page
	(*allDemos)[""] = Demo{
		Description: "Start page",
		Image:       "www/start.png",
		URL:         startPageURL,
	}

	return nil
}

func loadDemoAndSanitize(allDemos *map[string]Demo, configFile string) error {
	newDemos := make(map[string]Demo)

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("%s config doesn't exist: %v", configFile, err)
	}

	if err := yaml.Unmarshal(data, &newDemos); err != nil {
		return fmt.Errorf("%s isn't a valid yaml file: %v", configFile, err)
	}

	baseDir := filepath.Dir(configFile)

	// remove invalid elements, translate image paths relative to config file and set default timer
	for id, d := range newDemos {
		if d.URL == "" && len(d.Slides) == 0 {
			fmt.Printf("Removing %s has no url nor slides attributes\n", id)
			delete(newDemos, id)
		}
		if len(d.Slides) > 0 && d.Time == 0 {
			d.Time = defaultTime
		}

		// if relative image dir, prefix baseDir
		if d.Image != "" && !strings.HasPrefix(d.Image, "/") {
			d.Image = path.Join(baseDir, d.Image)
		}
		for i, s := range d.Slides {
			if s.Image != "" && !strings.HasPrefix(s.Image, "/") {
				s.Image = path.Join(baseDir, s.Image)
				d.Slides[i] = s
			}
		}

		// Take first slide image if no Image is set.
		if d.Image == "" && len(d.Slides) > 0 {
			d.Image = d.Slides[0].Image
		}
		if d.URL != "" && len(d.Slides) > 0 {
			fmt.Printf("%s has both url and slides attributes. Will only use slides\n", id)
			d.URL = ""
		}

		newDemos[id] = d
	}

	// copy from newDemos to allDemos. Override as later is always more specific
	for k, v := range newDemos {
		(*allDemos)[k] = v
	}

	return nil
}

func getValidDemosConfig(base string) ([]string, error) {
	var demoConfigs []string
	demoBaseDirs, err := ioutil.ReadDir(config.DemoBaseDir)
	if err != nil {
		return nil, fmt.Errorf("Can't read %s path", base)
	}
	for _, dir := range demoBaseDirs {
		p := path.Join(base, dir.Name(), "current", demoDefaultFilename)
		if _, err := os.Stat(p); err == nil {
			demoConfigs = append(demoConfigs, p)
		}
	}

	return demoConfigs, nil
}
