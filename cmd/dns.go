/*
Copyright 2024-2025 LinuxSuRen.

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
	"fmt"
	adns "github.com/linuxsuren/atest-ext-collector/pkg/dns"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func createDNSCmd() (cmd *cobra.Command) {
	opt := &dnsOptions{}
	cmd = &cobra.Command{
		Use:     "dns",
		Short:   "A simple DNS server",
		PreRunE: opt.preRunE,
		RunE:    opt.runE,
	}
	opt.setFlags(cmd.Flags())
	return
}

type dnsOptions struct {
	upstream     string
	port         int
	httpPort     int
	simpleConfig string
	cache        string

	// inner fields
	cacheHandler adns.DNSCache
	server       adns.Server
	config       *adns.DNSConfig
}

func (o *dnsOptions) setFlags(flags *pflag.FlagSet) {
	flags.StringVarP(&o.upstream, "upstream", "u", "8.8.8.8:53", "upstream dns server")
	flags.IntVarP(&o.port, "port", "p", 53, "The port for the dns server")
	flags.IntVarP(&o.httpPort, "http-port", "", 9090, "The port for the http server")
	flags.StringVarP(&o.simpleConfig, "simple-config", "", "", "A map based simple config of DNS records")
	flags.StringVarP(&o.cache, "cache", "", "memory", fmt.Sprintf("The DNS cache type, supported: %v", adns.GetDNSCacheNames()))
}

func (o *dnsOptions) preRunE(_ *cobra.Command, _ []string) (err error) {
	if o.cacheHandler = adns.GetDNSCache(o.cache); o.cacheHandler == nil {
		err = fmt.Errorf("cannot find cache: %s", o.cache)
		return
	}

	if o.simpleConfig != "" {
		if o.config, err = adns.ParseFromFile(o.simpleConfig); err != nil {
			return
		}

		o.cacheHandler.Init(o.config.Simple)
		o.cacheHandler.GetWildcardCache().Init(o.config.Wildcard)
	}
	if o.config == nil {
		o.config = &adns.DNSConfig{}
	}

	if o.upstream != "" {
		o.config.Upstream = o.upstream
	}
	if o.port > 0 {
		o.config.Port = o.port
	}
	o.server = adns.NewDNSServer(o.config, o.cacheHandler)
	return
}

func (o *dnsOptions) runE(cmd *cobra.Command, args []string) (err error) {
	go func() {
		if err := adns.NewHTTPServer(o.httpPort, o.upstream, o.cacheHandler).Start(); err != nil {
			panic(err)
		}
	}()

	err = o.server.Start()
	return
}
