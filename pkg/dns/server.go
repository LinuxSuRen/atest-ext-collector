/*
Copyright 2025 LinuxSuRen.

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

package dns

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/miekg/dns"
	"net"
	"strings"
)

type dnsServer struct {
	config       *DNSConfig
	cacheHandler DNSCache
}

func NewDNSServer(config *DNSConfig, cacheHandler DNSCache) Server {
	return &dnsServer{
		config:       config,
		cacheHandler: cacheHandler,
	}
}

func (d *dnsServer) Start() (err error) {
	addr := net.UDPAddr{
		Port: d.config.Port,
		IP:   net.ParseIP("0.0.0.0"),
	}
	var u *net.UDPConn
	if u, err = net.ListenUDP("udp", &addr); err != nil {
		return
	}

	fmt.Println("DNS server is ready! Port is:", d.config.Port)
	// Wait to get request on that port
	for {
		tmp := make([]byte, 1024)
		_, addr, _ := u.ReadFrom(tmp)
		clientAddr := addr
		packet := gopacket.NewPacket(tmp, layers.LayerTypeDNS, gopacket.Default)
		dnsPacket := packet.Layer(layers.LayerTypeDNS)
		tcp, _ := dnsPacket.(*layers.DNS)
		if !d.serveDNS(u, clientAddr, tcp) {
			d.cacheDNS(string(tcp.Questions[0].Name))
		}
	}
	return
}

func (d *dnsServer) serveDNS(u *net.UDPConn, clientAddr net.Addr, request *layers.DNS) (resolved bool) {
	replyMess := request
	var dnsAnswer layers.DNSResourceRecord
	dnsAnswer.Type = layers.DNSTypeA
	var ip string
	var err error
	resolved = true

	// check if this is a black domain
	domain := string(request.Questions[0].Name)
	if d.cacheHandler.IsBlackDomain(domain) {
		return
	} else {
		ip = d.cacheHandler.LookupIP(domain)
		if ip == "" {
			if ip = d.cacheHandler.GetWildcardCache().LookupIP(domain); ip == "" {
				resolved = false
				return
			}
		}
	}

	a, _, _ := net.ParseCIDR(ip + "/24")
	dnsAnswer.Type = layers.DNSTypeA
	dnsAnswer.IP = a
	dnsAnswer.Name = []byte(domain)
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

func (d *dnsServer) cacheDNS(name string) {
	client := dns.Client{}
	var m dns.Msg
	m.SetQuestion(name+".", dns.TypeA)
	reply, _, err := client.Exchange(&m, d.config.Upstream)
	if err != nil {
		fmt.Println("failed query domain", name, err)
		return
	}
	if reply.Rcode == dns.RcodeSuccess && len(reply.Answer) > 0 {
		if a, ok := reply.Answer[0].(*dns.A); ok {
			d.cacheHandler.Put(strings.TrimSuffix(reply.Question[0].Name, "."), a.A.String())
			fmt.Println("cache new record", reply.Question[0].Name, "==", a.A.String())
			fmt.Println("total cache item count", d.cacheHandler.Size())
			return
		} else if cname, ok := reply.Answer[0].(*dns.CNAME); ok {
			fmt.Println(name, "==", strings.TrimSuffix(cname.Hdr.Name, "."), "==", strings.TrimSuffix(cname.Target, "."))
			if ip := d.cacheHandler.LookupIP(strings.TrimSuffix(cname.Target, ".")); ip != "" {
				d.cacheHandler.Put(strings.TrimSuffix(cname.Hdr.Name, "."), ip)
			} else {
				d.cacheDNS(strings.TrimSuffix(cname.Target, "."))
			}
		}
		fmt.Println("unknown record", reply.Question[0].Name, ",", reply.Answer[0].String())
	}
}

func (d *dnsServer) Stop() (err error) {
	return
}
