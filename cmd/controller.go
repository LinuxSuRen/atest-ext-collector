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
	"github.com/spf13/pflag"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/elazarl/goproxy"
	"github.com/linuxsuren/atest-ext-collector/pkg"
	"github.com/spf13/cobra"
)

type proxyOption struct {
	port     int
	verbose  bool
	upstream string

	handler func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string)
}

type controllerOption struct {
	proxyOption
	ctrl *pkg.Controller
}

func createControllerCmd() (c *cobra.Command) {
	opt := &controllerOption{}
	c = &cobra.Command{
		Use:   "controller",
		Short: "HTTP network controller",
		RunE:  opt.runE,
		Args:  cobra.MinimumNArgs(1),
	}
	opt.SetFlags(c.Flags())
	return
}

func (o *proxyOption) SetFlags(flags *pflag.FlagSet) {
	flags.IntVarP(&o.port, "port", "p", 9090, "The port for the proxy")
	flags.BoolVarP(&o.verbose, "verbose", "", false, "Verbose mode")
	flags.StringVarP(&o.upstream, "upstream", "", "", "The upstream proxy")
}

func (o *controllerOption) readController(c *cobra.Command, args []string) (err error) {
	o.ctrl, err = pkg.ParseController(args[0])
	o.handler = o.ctrl.ConnectFilter
	return
}

func (o *proxyOption) runE(c *cobra.Command, args []string) (err error) {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = o.verbose
	if o.upstream != "" {
		proxy.Tr.Proxy = func(r *http.Request) (*url.URL, error) {
			return url.Parse(o.upstream)
		}
		proxy.ConnectDial = proxy.NewConnectDialToProxy(o.upstream)
		c.Println("Using upstream proxy", o.upstream)
	}
	proxy.OnRequest().HandleConnectFunc(o.handler)

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
