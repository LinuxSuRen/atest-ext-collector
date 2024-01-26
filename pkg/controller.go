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
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/elazarl/goproxy"
	"gopkg.in/yaml.v3"
)

type Controller struct {
	WhiteList []WhiteItem `yaml:"whiteList"`
	Windows   []Window    `yaml:"windows"`
}

type WhiteItem struct {
	Host     string        `yaml:"host"`
	Duration time.Duration `yaml:"duration"`
}

type Window struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

func (w Window) getWindowTime() (fromTime time.Time, toTime time.Time, err error) {
	now := time.Now()

	from, fErr := time.Parse("15:04", w.From)
	to, tErr := time.Parse("15:04", w.To)
	if fErr != nil || tErr != nil {
		err = fmt.Errorf("find wrong time format: %q, %q, error is: %v, %v", w.From, w.To, fErr, tErr)
		return
	}

	fromTime = time.Date(now.Year(), now.Month(), now.Day(),
		from.Hour(), from.Minute(), now.Second(), now.Nanosecond(), now.Location())
	toTime = time.Date(now.Year(), now.Month(), now.Day(),
		to.Hour(), to.Minute(), now.Second(), now.Nanosecond(), now.Location())
	return
}

func ParseController(config string) (ctrl *Controller, err error) {
	var data []byte
	if data, err = os.ReadFile(config); err == nil {
		ctrl = &Controller{}
		err = yaml.Unmarshal(data, ctrl)
	}
	return
}

func (c *Controller) ConnectFilter(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
	inWindow := false
	now := time.Now()

	for _, w := range c.Windows {
		if from, to, err := w.getWindowTime(); err != nil {
			log.Printf("%v", err)
			continue
		} else if now.After(from) && now.Before(to) {
			inWindow = true
			break
		}
	}

	if !inWindow {
		log.Printf("reject: %q due to out of window\n", host)
		return goproxy.RejectConnect, host
	}

	for _, w := range c.WhiteList {
		ok, err := regexp.MatchString(w.Host, host)
		if err != nil {
			log.Printf("find wrong pattern: %q, error is: %v", w.Host, err)
		} else if ok {
			return goproxy.OkConnect, host
		}
	}

	log.Printf("reject: %q\n", host)
	return goproxy.RejectConnect, host
}
