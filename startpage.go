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
	"fmt"
	"html/template"
	"net/http"
	"path"

	"github.com/ubuntu/booth-demo-manager/config"
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
		data.Addrs = append(data.Addrs, "We couldn't find any other network IP on this network configuration.")
		ips = append(ips, "localhost")
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
