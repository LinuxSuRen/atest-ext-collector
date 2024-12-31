package cmd

import (
	"github.com/linuxsuren/atest-ext-collector/pkg/gateway"
	"github.com/spf13/cobra"
)

func createGatewayCmd() (cmd *cobra.Command) {
	opt := &gatewayOptions{}
	cmd = &cobra.Command{
		Use:  "gateway",
		RunE: opt.runE,
	}
	opt.setupFlags(cmd)
	return
}

type gatewayOptions struct {
	certFile string
	keyFile  string
}

func (o *gatewayOptions) setupFlags(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.StringVarP(&o.certFile, "cert", "c", "", "The cert file")
	flags.StringVarP(&o.keyFile, "key", "k", "", "The key file")
}

func (o *gatewayOptions) runE(cmd *cobra.Command, args []string) (err error) {
	var gw *gateway.Gateway
	if gw, err = gateway.ParseGateway(args[0]); err != nil {
		return
	}
	server := gateway.GatewayServer{}
	server.WithGateway(gw)
	server.WithTLS(o.certFile, o.keyFile)
	server.Start()
	return
}
