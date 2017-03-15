# Booth Demo manager
This application enables managing multiple websites by piloting a webpage.

Serve a piloted webpage (default port being *8001*) and a pilot interface
to control it.

You  can either switch from demo to demo, or have complex demo with multiple
pages, cycling after a defined time between them.

If you want to point out a particular page or slide of one demo, you can as
well easily stop the timer by clicking on a particular element.

## Usage
Just start `boot-demo-manager`, it will serve 2 web interfaces:
* One at http://localhost:8001/ (the frontend user to display), which
is the piloted page.
* One at http://localhost:8001/pilot, which shows all available demos and
pages for complex demos, and enables switching display frontend between them.

Note that multiple display and pilot instances are possible. They will all be
in sync.

The port can be changed via the *-p* option.

## Configuration

### Detection
Configuration is a mix and match of base definition and autodetection (config
shipped by some snaps). It will look for `booth-demo-manager.def` files in:
* Code directory (the snap doesn't ship any config file though)
* Autodetected snaps: any snap shipping a `booth-demo-manager.def` in its root
directory.
* *booth-demo-manager* snap data directory (`SNAP_DATA`) containing a `booth-demo-manager.def`.
* Relative local path (for master only) with `booth-demo-manager.def` in root

If multiple configurations have the same demo key in a later config file layer, it
will override the definition in a previous one.

We can disable this detection logic, providing manually a definition file format via the *-c* option.

The path to this file can be as well defined via the *-c* option.

### Definition file format

The definition file format is a yaml file.

#### A single website
```
ubuntu:
  description: "Ubuntu website"
  image: "images/ubuntu.svg"
  url: https://www.ubuntu.com
```
- The key is an ID.
- *description* will show off in the /pilot page
- *image* is an optional image path for the /pilot page. This can be an absolute path, or relative to the `.def` file.
- *url* is the url where we'll be navigated to when selecting this demo.

#### A slide deck

```
ubuntu-core-slides:
  description: "Ubuntu core slide deck presentation"
  time: 45
  slides:
    - image: "img/slide1_preview.png"
      url: http://localhost:9001/slide1.svg
    - image: "img/slide2_preview.png"
      url: http://localhost:9001/slide2.svg
    - image: "img/slide3_preview.png"
      url: http://localhost:9001/slide3.svg
```

The slide deck differs from our previous use case that it has multiple urls it can navigate to
and will loop over them automatically.
- The key is an ID.
- *description* will show off in the /pilot page
- *time* is an optional time integer in seconds which is the delay until next slide is being displayed. Default is 30 (seconds).
- then, *slides* keyword is followed by tuples of:
  - *image*, optional image path for the /pilot page. This can be an absolute path, or relative to the `.def` file.
  - *url* is the url where we'll be navigated to when selecting this demo.

#### Example
Have a look at some configuration example in [`examples/booth-demo-manager.def`](./examples/booth-demo-manager.def).
