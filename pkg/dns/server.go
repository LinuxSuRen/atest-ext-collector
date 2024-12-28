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
	dnsCache DNSCache
}

func NewHTTPServer(port int, dnsCache DNSCache) Server {
	return &httpServer{
		port:     port,
		dnsCache: dnsCache,
	}
}

func (s *httpServer) Start() (err error) {
	var listener net.Listener
	listener, err = net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.home)
	mux.HandleFunc("/remove", s.removeData)
	mux.HandleFunc("/add", s.addData)

	server := &http.Server{
		Handler: mux,
	}
	err = server.Serve(listener)
	return
}

func (s *httpServer) home(w http.ResponseWriter, r *http.Request) {
	tpl, err := template.New("DNS Simple Server").Parse(string(frontPage))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tpl.Execute(w, map[string]interface{}{
		"cache": s.dnsCache.Data(),
		"size":  s.dnsCache.Size(),
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
	r.ParseForm()
	domain := r.Form.Get("domain")
	ip := r.Form.Get("ip")
	s.dnsCache.Put(domain, ip)
	http.Redirect(w, r, "/", http.StatusPermanentRedirect)
}

func (s *httpServer) Stop() (err error) {
	return
}

//go:embed data/index.html
var frontPage []byte
