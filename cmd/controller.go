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

package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/elazarl/goproxy"
	"github.com/linuxsuren/atest-ext-collector/pkg"
	"github.com/spf13/cobra"
)

type controllerOption struct {
	port     int
	verbose  bool
	upstream string
}

func createControllerCmd() (c *cobra.Command) {
	opt := &controllerOption{}
	c = &cobra.Command{
		Use:   "controller",
		Short: "HTTP network controller",
		RunE:  opt.runE,
		Args:  cobra.MinimumNArgs(1),
	}
	flags := c.Flags()
	flags.IntVarP(&opt.port, "port", "p", 9090, "The port for the proxy")
	flags.BoolVarP(&opt.verbose, "verbose", "", false, "Verbose mode")
	flags.StringVarP(&opt.upstream, "upstream", "", "", "The upstream proxy")
	return
}

func (o *controllerOption) runE(c *cobra.Command, args []string) (err error) {
	var ctrl *pkg.Controller
	if ctrl, err = pkg.ParseController(args[0]); err != nil {
		return
	}

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = o.verbose
	if o.upstream != "" {
		proxy.Tr.Proxy = func(r *http.Request) (*url.URL, error) {
			return url.Parse(o.upstream)
		}
		proxy.ConnectDial = proxy.NewConnectDialToProxy(o.upstream)
		c.Println("Using upstream proxy", o.upstream)
	}
	proxy.OnRequest().HandleConnectFunc(ctrl.ConnectFilter)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", o.port),
		Handler: proxy,
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		_ = srv.Shutdown(context.Background())
	}()

	c.Println("Starting the proxy server with port", o.port)
	_ = srv.ListenAndServe()
	return
}
