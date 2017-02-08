package pilot

import "time"

// Demo represent one demo element which can be a slide deck or multiple items
type Demo struct {
	Description string
	Image       string
	Time        int
	URL         string `yaml:"url"`
	Slides      []struct {
		Image string
		URL   string `yaml:"url"`
	}
}

// CurrentDemo is an augmented Demo item with metadata on current slides, ID and such
type CurrentDemo struct {
	*Demo
	id         string
	url        string
	slideIndex int
	ticker     *time.Ticker
	stop       chan struct{}
	deselected chan bool
}

// Select set demo as current demo. Starts timer if slides demo and not already the case and set
// index as current one. It will spawn a goroutine to send some change current after timer.
func (d Demo) Select(id string, slideIndex int, currentChan chan<- CurrentDemoMsg) *CurrentDemo {
	if !d.IsSlideDemo() {
		c := &CurrentDemo{&d, id, d.URL, slideIndex, nil, nil, make(chan bool)}
		sendNewCurrentURL(currentChan, c)
		return c
	}

	// non auto rolling demo: only select one slide
	if slideIndex > -1 && slideIndex < len(d.Slides) {
		url := d.Slides[slideIndex].URL
		c := &CurrentDemo{&d, id, url, slideIndex, nil, nil, make(chan bool)}
		sendNewCurrentURL(currentChan, c)
		return c
	}

	// sliding demo
	ticker := time.NewTicker(time.Second * time.Duration(d.Time))
	stop := make(chan struct{})
	// First immediately first elem
	c := &CurrentDemo{&d, id, d.Slides[0].URL, 0, ticker, stop, make(chan bool)}
	sendNewCurrentURL(currentChan, c)
	go func() {
		defer func() { current.deselected <- true }()
		for {
			select {
			case <-c.ticker.C:
				c.slideIndex = (c.slideIndex + 1) % len(c.Slides)
				c.url = d.Slides[c.slideIndex].URL
				sendNewCurrentURL(currentChan, c)
			case <-c.stop:
				c.ticker.Stop()
				return
			}
		}
	}()
	return c
}

// Release teardown current demo as current one. It's only cleaning the ticker goroutine if it was
// a slide demo.
func (c *CurrentDemo) Release() {
	// No need to tear down ticker if none (none slide demo or fixed selection)
	if !c.IsSlideDemo() || c.stop == nil {
		go func() {
			c.deselected <- true
		}()
		return
	}
	close(c.stop)
}

// IsSlideDemo returns the nature of the demo (slide with auto-advance or fixed demo)
func (d Demo) IsSlideDemo() bool {
	return len(d.Slides) != 0
}
