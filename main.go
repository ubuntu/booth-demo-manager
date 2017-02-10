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
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"path"
	"strings"

	"github.com/ubuntu/booth-demo-manager/config"
)

var (
	port *string
)

func main() {
	port = flag.String("p", "8001", "port on which to serve web interfaces")

	flag.Parse()

	// Initialize websocket communication servers
	initWS()

	// start pilot system
	if err := startPilot(); err != nil {
		log.Fatalf("Couldn't load demo settings: %v", err)
	}

	wwwPath := path.Join(config.Rootdir, "www")

	// Generated links: will serve IP to connect to
	http.HandleFunc("/start", startPageHandler)
	http.HandleFunc("/pilot/demos/", func(w http.ResponseWriter, r *http.Request) {
		// Serve Index pilot page for generated /demo/ links. The client-side router will handle it then.
		http.ServeFile(w, r, path.Join(wwwPath, "pilot", "index.html"))
	})

	// Website real assets
	wwwHandler := http.FileServer(http.Dir(path.Join(config.Rootdir, "www")))
	dirs, err := ioutil.ReadDir(wwwPath)
	if err != nil {
		log.Fatal("Couldn't list content of ", wwwPath)
	}
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		// display is /
		if dir.Name() == "display" {
			http.Handle("/", http.FileServer(http.Dir(path.Join(config.Rootdir, "www", "display"))))
			continue
		}
		http.Handle("/"+dir.Name()+"/", wwwHandler)
	}
	// Virtual image directory
	http.Handle("/pilot/generatedimg/", http.StripPrefix("/pilot/generatedimg/", http.FileServer(VirtImages{})))

	// Print ips
	ips, err := getLocalIPs()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Serving display on http://localhost:%s", *port)
	for _, ip := range ips {
		addr := fmt.Sprintf("http://%s:%s/pilot", ip, *port)
		log.Printf("You access pilot interface via %s\n", addr)
	}

	// Start server
	if err = http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatal("Couldn't start webserver:", err)
	}

}

func getLocalIPs() (ips []string, err error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("Getting network IP error: %v", err)
	}
	for _, addr := range addrs {
		ip := strings.Split(addr.String(), "/")[0]
		if ip == "127.0.0.1" || strings.HasPrefix(ip, ":") {
			continue
		}
		ips = append(ips, ip)
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("No local network found for starting pilot web interface")
	}
	return ips, nil
}
