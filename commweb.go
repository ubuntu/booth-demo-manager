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
	"log"

	"github.com/ubuntu/booth-demo-manager/messages"
	"github.com/ubuntu/booth-demo-manager/pilot"
)

var (
	allDemos               map[string]pilot.Demo
	current                pilot.CurrentDemoMsg
	displayComm, pilotComm *messages.Server
)

func initWS() {
	// Websocket servers
	displayComm = messages.NewServer("/api/display", newDisplayClient)
	//defer displayComm.Quit()
	go displayComm.Listen()
	pilotComm = messages.NewServer("/api/pilot", newPilotClient)
	//defer pilotComm.Quit()
	go pilotComm.Listen()
}

func buildCurrentDemoMessage(currDemo pilot.CurrentDemoMsg) *messages.Action {
	return &messages.Action{
		Command: "current",
		Content: currDemo,
	}
}

func buildAllDemosChangedMessage(allDemos map[string]pilot.Demo) *messages.Action {
	return &messages.Action{
		Command: "allDemos",
		Content: allDemos,
	}
}

func newDisplayClient(c *messages.Client) {
	if current.ID == "" {
		return
	}
	c.Send(buildCurrentDemoMessage(current))
}

func newPilotClient(c *messages.Client) {
	if len(allDemos) == 0 {
		return
	}
	c.Send(buildAllDemosChangedMessage(allDemos))
	if current.ID == "" {
		return
	}
	c.Send(buildCurrentDemoMessage(current))
}

func startPilot() error {
	changeCurrentDemo := make(chan pilot.CurrentDemoMsg)
	currDemoChanged, allDemosChanged, err := pilot.Start(changeCurrentDemo)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case a := <-pilotComm.Messages:
				switch a.Command {
				case "changeCurrent":
					// TODO: automated conversion/duck typing?
					orig, ok := a.Content.(map[string]interface{})
					if !ok {
						log.Println("Badly formatted websocket request:", a.Content)
						continue
					}
					// We just skip unmatched values
					id, _ := orig["ID"].(string)
					index, _ := orig["Index"].(float64)
					url, _ := orig["URL"].(string)
					newCurr := pilot.CurrentDemoMsg{
						ID:    id,
						Index: int(index),
						URL:   url,
					}
					changeCurrentDemo <- newCurr
				}
			case curr := <-currDemoChanged:
				current = curr
				msg := buildCurrentDemoMessage(curr)
				displayComm.Send(msg)
				pilotComm.Send(msg)
			case all := <-allDemosChanged:
				// build demo map
				allDemos = all
				msg := buildAllDemosChangedMessage(allDemos)
				pilotComm.Send(msg)
			}
		}
	}()
	return nil
}
