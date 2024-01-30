/*
Copyright 2024 LinuxSuRen.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pkg

import (
	"encoding/json"
	"io"
	"net/http"
)

type ControllerRest struct {
	ConfigFile string
}

func NewControllerRest(configFile string) *ControllerRest {
	return &ControllerRest{
		ConfigFile: configFile,
	}
}

func (c *ControllerRest) GetConfig(w http.ResponseWriter, req *http.Request) {
	ctrl, err := c.getController()
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(err.Error()))
	} else {
		w.Write(ctrl.ToJSONData())
	}
}

func (c *ControllerRest) AddWhiteItem(w http.ResponseWriter, req *http.Request) {
	ctrl, err := c.getController()
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	data, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	item := &WhiteItem{}
	if err = json.Unmarshal(data, item); err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	if !item.IsValid() {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("white item is invalid"))
		return
	}

	duplicated := false
	for _, i := range ctrl.WhiteList {
		if i.Host == item.Host {
			duplicated = true
			break
		}
	}

	if !duplicated {
		ctrl.WhiteList = append(ctrl.WhiteList, *item)
	}

	if err = SaveController(c.ConfigFile, ctrl); err != nil {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("failed to save controller"))
		return
	}

	_, _ = w.Write([]byte("ok"))
}

func (c *ControllerRest) DelWhiteItem(w http.ResponseWriter, req *http.Request) {}

func (c *ControllerRest) AddWindowItem(w http.ResponseWriter, req *http.Request) {}

func (c *ControllerRest) DelWindowItem(w http.ResponseWriter, req *http.Request) {}

func (c *ControllerRest) getController() (*Controller, error) {
	return ParseController(c.ConfigFile)
}
