package pilot

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"log"

	"time"

	yaml "gopkg.in/yaml.v2"
)

// CurrentDemo element with ID, its current URL and optional index
type CurrentDemo struct {
	ID    string
	Index int
	URL   string
}

var (
	allDemos map[string]Demo
	current  CurrentDemo
)

// Demo represent one demo element which can be a slide deck or multiple items
type Demo struct {
	Description string
	Image       string
	Time        int
	URL         string `yaml:"url"`
	Slides      []struct {
		Image string
		URLS  string `yaml:"urls"`
	}
}

const (
	demoFilename = "demos.def"
	defaultTime  = 30
)

// Start the pilot element handling timers and such. Return a channel of current demo ID
// and all demos
// TODO: starts and close it properly once we can shutdown webserver
func Start(changeCurrent <-chan CurrentDemo) (<-chan CurrentDemo, <-chan map[string]Demo, error) {
	currentCh := make(chan CurrentDemo)
	allDemosCh := make(chan map[string]Demo)

	if err := loadDefinition(); err != nil {
		return nil, nil, err
	}

	go func() {
		// sending first all Demos list
		allDemosCh <- allDemos

		var ticker *time.Ticker
		quitTicker := make(chan bool)
		defer close(quitTicker)
		for {
			select {
			case elem := <-changeCurrent:
				d, ok := allDemos[elem.ID]
				if !ok {
					log.Printf("%s not in currently available demos", elem.ID)
					continue
				}
				if ticker != nil {
					quitTicker <- true
				}

				url := d.URL
				// Handling demo with multiple URLs
				if len(d.Slides) > 0 {
					/*
						// select url to show
						url := d.URLS[0]
						for _, u := range d.URLS[0] {
							if u == elem.URL {
								// Found!
							}
						}

						ticker = time.NewTicker(time.Second * time.Duration(d.Time))
						go func() {
							for {
								select {
								case <-ticker.C:
								case <-quitTicker:
									ticker.Stop()
									ticker = nil
									return
								}
							}
						}()*/
				}
				current = CurrentDemo{ID: elem.ID, URL: url, Index: 0}

				currentCh <- current
			}
		}

	}()

	return currentCh, allDemosCh, nil
}

func loadDefinition() error {
	var data []byte
	var err error
	var selectedFile string
	for _, selectedFile := range []string{path.Join(datadir, demoFilename), path.Join(rootdir, demoFilename)} {
		data, err = ioutil.ReadFile(selectedFile)
		if err != nil {
			continue
		}
	}
	if data == nil {
		return fmt.Errorf("Couldn't read any of %s: %v", demoFilename, err)
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
			fmt.Printf("%s has both url nor slides attributes. Will only use slides\n", id)
		}
	}

	return nil
}

// FIXME: check how to share main Rootdir and Datadir
var (
	rootdir string
	datadir string
)

func init() {
	// Set main set of directories
	var err error
	rootdir = os.Getenv("SNAP")
	if rootdir == "" {
		if rootdir, err = filepath.Abs("."); err != nil {
			log.Fatal(err)
		}
	}
	datadir = os.Getenv("SNAP_DATA")
	if datadir == "" {
		datadir = rootdir
	}
}
