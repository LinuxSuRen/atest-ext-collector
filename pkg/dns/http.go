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

package dns

import (
	_ "embed"
	"fmt"
	"html/template"
	"net"
	"net/http"
)

type httpServer struct {
	port     int
	upstream string
	listener net.Listener
	dnsCache DNSCache
}

func NewHTTPServer(port int, upstream string, dnsCache DNSCache) Server {
	return &httpServer{
		port:     port,
		upstream: upstream,
		dnsCache: dnsCache,
	}
}

func (s *httpServer) Start() (err error) {
	s.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.home)
	mux.HandleFunc("/remove", s.removeData)
	mux.HandleFunc("/add", s.addData)
	mux.HandleFunc("/addBlack", s.addBlack)
	mux.HandleFunc("/removeBlack", s.removeBlack)

	server := &http.Server{
		Handler: mux,
	}
	err = server.Serve(s.listener)
	return
}

func (s *httpServer) home(w http.ResponseWriter, r *http.Request) {
	tpl, err := template.New("DNS Simple Server").Parse(string(frontPage))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tpl.Execute(w, map[string]interface{}{
		"cache":    s.dnsCache.Data(),
		"size":     s.dnsCache.Size(),
		"upstream": s.upstream,
		"black":    s.dnsCache.ListBlackDomains(),
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *httpServer) removeData(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	s.dnsCache.Remove(domain)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (s *httpServer) addData(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	domain := r.Form.Get("domain")
	ip := r.Form.Get("ip")
	s.dnsCache.Put(domain, ip)
	http.Redirect(w, r, "/", http.StatusPermanentRedirect)
}

func (s *httpServer) addBlack(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	s.dnsCache.AddBlackDomain(domain)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (s *httpServer) removeBlack(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	s.dnsCache.RemoveBlackDomain(domain)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (s *httpServer) Stop() (err error) {
	if s.listener != nil {
		err = s.listener.Close()
	}
	return
}

//go:embed data/index.html
var frontPage []byte
