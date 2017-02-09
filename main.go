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
	"github.com/ubuntu/booth-demo-manager/pilot"
)

var (
	port *string
)

func main() {
	port = flag.String("p", "8001", "port on which to serve web interfaces")
	pilot.DemoFilePath = flag.String("c", pilot.DemoDefaultFilename, "config file path overriding default one")

	flag.Parse()

	// Initialize websocket communication servers
	initWS()

	// start pilot system
	if err := startPilot(); err != nil {
		log.Fatalf("Couldn't load demo settings: %v", err)
	}

	// Website real assets
	wwwPath := path.Join(config.Rootdir, "www")
	wwwHandler := http.FileServer(http.Dir(path.Join(config.Rootdir, "www")))
	dirs, err := ioutil.ReadDir(wwwPath)
	if err != nil {
		log.Fatal("Couldn't list content of ", wwwPath)
	}
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		http.Handle("/"+dir.Name()+"/", wwwHandler)
	}
	// Virtual image directory
	http.Handle("/pilot/generatedimg/", http.StripPrefix("/pilot/generatedimg/", http.FileServer(VirtImages{})))
	// Generated links: will serve IP to connect to
	http.HandleFunc("/", startPageHandler)

	// Print ips
	ips, err := getLocalIPs()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Serving display on http://localhost:%s/display", *port)
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
