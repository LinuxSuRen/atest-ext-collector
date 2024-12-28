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
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	adns "github.com/linuxsuren/atest-ext-collector/pkg/dns"
	"github.com/miekg/dns"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
	"net"
	"os"
	"strings"
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
	simpleConfig string
	cache        string

	// inner fields
	cacheHandler adns.DNSCache
}

func (o *dnsOptions) setFlags(flags *pflag.FlagSet) {
	flags.StringVarP(&o.upstream, "upstream", "u", "8.8.8.8:53", "upstream dns server")
	flags.IntVarP(&o.port, "port", "p", 53, "The port for the dns server")
	flags.StringVarP(&o.simpleConfig, "simple-config", "", "", "A map based simple config of DNS records")
	flags.StringVarP(&o.cache, "cache", "", "memory", fmt.Sprintf("The DNS cache type, supported: %v", adns.GetDNSCacheNames()))
}

func (o *dnsOptions) preRunE(_ *cobra.Command, _ []string) (err error) {
	if o.cacheHandler = adns.GetDNSCache(o.cache); o.cacheHandler == nil {
		err = fmt.Errorf("cannot find cache: %s", o.cache)
		return
	}

	if o.simpleConfig != "" {
		var data []byte
		if data, err = os.ReadFile(o.simpleConfig); err == nil {
			records := make(map[string]string)
			if err = yaml.Unmarshal(data, &records); err == nil {
				o.cacheHandler.Init(records)
			}
		}
	}
	return
}

func (o *dnsOptions) runE(cmd *cobra.Command, args []string) (err error) {
	addr := net.UDPAddr{
		Port: o.port,
		IP:   net.ParseIP("0.0.0.0"),
	}
	var u *net.UDPConn
	if u, err = net.ListenUDP("udp", &addr); err != nil {
		return
	}

	cmd.Println("DNS server is ready!")
	// Wait to get request on that port
	for {
		tmp := make([]byte, 1024)
		_, addr, _ := u.ReadFrom(tmp)
		clientAddr := addr
		packet := gopacket.NewPacket(tmp, layers.LayerTypeDNS, gopacket.Default)
		dnsPacket := packet.Layer(layers.LayerTypeDNS)
		tcp, _ := dnsPacket.(*layers.DNS)
		if !o.serveDNS(u, clientAddr, tcp) {
			o.cacheDNS(string(tcp.Questions[0].Name))
		}
	}
	return
}

func (o *dnsOptions) cacheDNS(name string) {
	client := dns.Client{}
	var m dns.Msg
	m.SetQuestion(name+".", dns.TypeA)
	reply, _, err := client.Exchange(&m, o.upstream)
	if err != nil {
		fmt.Println("failed query domain", name, err)
		return
	}
	if reply.Rcode == dns.RcodeSuccess && len(reply.Answer) > 0 {
		if a, ok := reply.Answer[0].(*dns.A); ok {
			o.cacheHandler.Put(strings.TrimSuffix(reply.Question[0].Name, "."), a.A.String())
			fmt.Println("cache new record", reply.Question[0].Name, "==", a.A.String())
			fmt.Println("total cache item count", o.cacheHandler.Size())
			return
		} else if cname, ok := reply.Answer[0].(*dns.CNAME); ok {
			fmt.Println(name, "==", strings.TrimSuffix(cname.Hdr.Name, "."), "==", strings.TrimSuffix(cname.Target, "."))
			if ip := o.cacheHandler.LookupIP(strings.TrimSuffix(cname.Target, ".")); ip != "" {
				o.cacheHandler.Put(strings.TrimSuffix(cname.Hdr.Name, "."), ip)
			} else {
				o.cacheDNS(strings.TrimSuffix(cname.Target, "."))
			}
		}
		fmt.Println("unknown record", reply.Question[0].Name, ",", reply.Answer[0].String())
	}
}

func (o *dnsOptions) serveDNS(u *net.UDPConn, clientAddr net.Addr, request *layers.DNS) (resolved bool) {
	replyMess := request
	var dnsAnswer layers.DNSResourceRecord
	dnsAnswer.Type = layers.DNSTypeA
	var ip string
	var err error
	resolved = true
	ip = o.cacheHandler.LookupIP(string(request.Questions[0].Name))
	if ip == "" {
		//fmt.Printf("cannot found: %s\n", request.Questions[0].Name)
		resolved = false
		return
		//Todo: Log no data present for the IP and handle:todo
	}
	a, _, _ := net.ParseCIDR(ip + "/24")
	dnsAnswer.Type = layers.DNSTypeA
	dnsAnswer.IP = a
	dnsAnswer.Name = []byte(request.Questions[0].Name)
	fmt.Println(string(request.Questions[0].Name), "===", ip)
	dnsAnswer.Class = layers.DNSClassIN
	replyMess.QR = true
	replyMess.ANCount = 1
	replyMess.OpCode = layers.DNSOpCodeNotify
	replyMess.AA = true
	replyMess.Answers = append(replyMess.Answers, dnsAnswer)
	replyMess.ResponseCode = layers.DNSResponseCodeNoErr
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{} // See SerializeOptions for more details.
	err = replyMess.SerializeTo(buf, opts)
	if err != nil {
		panic(err)
	}
	u.WriteTo(buf.Bytes(), clientAddr)
	return
}
