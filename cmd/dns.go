package cmd

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/miekg/dns"
	"github.com/spf13/cobra"
	"net"
	"strings"
)

func createDNSCmd() (cmd *cobra.Command) {
	opt := &dnsOptions{}
	cmd = &cobra.Command{
		Use:  "dns",
		RunE: opt.runE,
	}
	return
}

type dnsOptions struct {
}

var records map[string]string

func (o *dnsOptions) runE(cmd *cobra.Command, args []string) (err error) {
	records = map[string]string{
		"baidu.com":     "223.143.166.121",
		"www.baidu.com": "223.143.166.121",
		"github.com":    "79.52.123.201",
		"linux.com":     "223.143.166.121",
		"www.linux.com": "223.143.166.121",
		"atest.com":     "127.0.0.1",
		"www.atest.com": "127.0.0.1",
	}

	addr := net.UDPAddr{
		Port: 53,
		IP:   net.ParseIP("127.0.0.1"),
	}
	var u *net.UDPConn
	if u, err = net.ListenUDP("udp", &addr); err != nil {
		return
	}

	// Wait to get request on that port
	for {
		tmp := make([]byte, 1024)
		_, addr, _ := u.ReadFrom(tmp)
		clientAddr := addr
		packet := gopacket.NewPacket(tmp, layers.LayerTypeDNS, gopacket.Default)
		dnsPacket := packet.Layer(layers.LayerTypeDNS)
		tcp, _ := dnsPacket.(*layers.DNS)
		if !serveDNS(u, clientAddr, tcp) {
			cacheDNS(string(tcp.Questions[0].Name))
		}
	}
	return
}

func cacheDNS(name string) {
	client := dns.Client{}
	var m dns.Msg
	m.SetQuestion(name+".", dns.TypeA)
	reply, _, err := client.Exchange(&m, "8.8.8.8:53")
	if err != nil {
		fmt.Println("failed query domain", name, err)
		return
	}
	if reply.Rcode == dns.RcodeSuccess && len(reply.Answer) > 0 {
		if a, ok := reply.Answer[0].(*dns.A); ok {
			records[strings.TrimSuffix(reply.Question[0].Name, ".")] = a.A.String()
			fmt.Println("cache new record", reply.Question[0].Name, "==", a.A.String())
			fmt.Println("total cache item count", len(records))
			return
		} else if cname, ok := reply.Answer[0].(*dns.CNAME); ok {
			fmt.Println(name, "==", strings.TrimSuffix(cname.Hdr.Name, "."), "==", strings.TrimSuffix(cname.Target, "."))
			if ip, ok := records[strings.TrimSuffix(cname.Target, ".")]; ok {
				records[strings.TrimSuffix(cname.Hdr.Name, ".")] = ip
			} else {
				cacheDNS(strings.TrimSuffix(cname.Target, "."))
			}
		}
		fmt.Println("unknown record", reply.Question[0].Name, ",", reply.Answer[0].String())
	}
}

func serveDNS(u *net.UDPConn, clientAddr net.Addr, request *layers.DNS) (resolved bool) {
	replyMess := request
	var dnsAnswer layers.DNSResourceRecord
	dnsAnswer.Type = layers.DNSTypeA
	var ip string
	var err error
	var ok bool
	resolved = true
	ip, ok = records[string(request.Questions[0].Name)]
	if !ok {
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
