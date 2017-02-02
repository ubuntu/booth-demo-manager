package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var (
	// Rootdir executable code to reach assets
	Rootdir string
	// Datadir access to write storage path
	Datadir string

	port *string
)

func init() {
	// Set main set of directories
	var err error
	Rootdir = os.Getenv("SNAP")
	if Rootdir == "" {
		if Rootdir, err = filepath.Abs("."); err != nil {
			log.Fatal(err)
		}
	}
	Datadir = os.Getenv("SNAP_DATA")
	if Datadir == "" {
		Datadir = Rootdir
	}

}

func main() {
	port = flag.String("p", "8100", "port to serve display interface")

	flag.Parse()

	ips, err := getLocalIPs()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Serving display on http://localhost:%s/display", *port)
	for _, ip := range ips {
		addr := fmt.Sprintf("http://%s:%s/pilot", ip, *port)
		log.Printf("You access pilot interface via %s\n", addr)
	}

	wwwHandler := http.FileServer(http.Dir(path.Join(Rootdir, "www")))
	// website assets
	http.Handle("/display/", wwwHandler)
	http.Handle("/pilot/", wwwHandler)
	http.Handle("/css/", wwwHandler)
	// generated links: will serve IP to connect to
	http.HandleFunc("/", startPageHandler)
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
