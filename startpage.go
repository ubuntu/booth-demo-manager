package main

import (
	"fmt"
	"html/template"
	"net/http"
	"path"

	"github.com/ubuntu/display-snap/config"
)

type startPageData struct {
	Addrs []string
}

func startPageHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Addrs []string
	}{}

	// get IPs to file up data
	ips, err := getLocalIPs()
	if err != nil {
		http.Error(w, fmt.Sprintf("Couldn't find any local IP on this device: %v", err), http.StatusInternalServerError)
		return
	}
	for _, ip := range ips {
		addr := fmt.Sprintf("http://%s:%s/pilot", ip, *port)
		data.Addrs = append(data.Addrs, addr)
	}

	t, err := template.ParseFiles(path.Join(config.Rootdir, "www", "start.html.tpl"))
	if err != nil {
		http.Error(w, fmt.Sprintf("Couldn't find starting page: %v", err), http.StatusInternalServerError)
		return
	}
	t.Execute(w, data)
}
