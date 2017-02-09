# Booth Demo manager
This application enables managing multiple websites by piloting a webpage.

Serve a piloted webpage (default port being 8001) and a pilot interface
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
Slides configuration is named `demos.def` in the current directory.
The path to this file can be as well defined via the *-c* option.

Have a look at some configuration example from [`demos.def.example`](./demos.def.example) in the
current directory.
