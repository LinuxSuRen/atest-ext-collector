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
	"github.com/linuxsuren/atest-ext-collector/pkg"
	"github.com/spf13/cobra"
)

func createProxyCmd() (cmd *cobra.Command) {
	ctr := &pkg.Controller{
		Windows: []pkg.Window{{
			From: "00:00",
			To:   "23:59",
		}},
		WhiteList: []pkg.FilterItem{{
			Host: ".*",
		}},
		BlackList: []pkg.FilterItem{{
			Host: ".*.stgowan.com",
		}, {
			Host: ".*.nanlanling.com",
		}, {
			Host: "cbjs.baidu.com",
		}, {
			Host: ".*.cnzz.com",
		}, {
			Host: ".*.bdstatic.com",
		}, {
			Host: ".*.qhimg.com",
		}, {
			Host: ".*.volces.com",
		}},
	}

	opt := &proxyOption{
		handler: ctr.ConnectFilter,
	}
	cmd = &cobra.Command{
		Use:   "proxy",
		Short: "HTTP proxy server",
		RunE:  opt.runE,
	}
	opt.SetFlags(cmd.Flags())
	return
}
