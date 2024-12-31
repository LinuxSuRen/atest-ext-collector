package gateway

import (
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type GatewayServer struct {
	Router   *mux.Router
	Gateway  *Gateway
	certFile string
	keyFile  string
}

func (g *GatewayServer) WithGateway(gw *Gateway) {
	g.Gateway = gw
}

func (g *GatewayServer) WithTLS(certFile, keyFile string) {
	g.certFile = certFile
	g.keyFile = keyFile
}

func (g *GatewayServer) Start() {
	g.Router = mux.NewRouter()

	g.Router.HandleFunc("/{any:.*}", func(writer http.ResponseWriter, r *http.Request) {
		fmt.Println(r.RequestURI)
		if g.Gateway == nil {
			return
		}

		for _, server := range g.Gateway.Servers {
			if server.Domain == r.Host {

				for _, route := range server.Routes {
					fmt.Println(r.URL.Path, route.Path)
					if !strings.HasPrefix(r.URL.Path, route.Path) {
						continue
					}

					targetURL, err := url.Parse(route.ProxyPass + strings.TrimSuffix(r.URL.Path, route.Path))
					for k, v := range r.URL.Query() {
						targetURL.Query().Set(k, strings.Join(v, ","))
					}
					newRequest, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL.RequestURI(), r.Body)
					//if err := r.ParseForm(); err == nil {
					//	newRequest.Form = url.Values{}
					//	for k, v := range r.Form {
					//		newRequest.Form[k] = v
					//	}
					//	newRequest.PostForm = url.Values{}
					//	for k, v := range r.PostForm {
					//		newRequest.PostForm[k] = v
					//	}
					//} else {
					//	log.Println("failed to parse form", err)
					//}
					newRequest.Header = r.Header

					fmt.Println("proxy pass the new request", targetURL.RequestURI())
					resp, err := http.DefaultClient.Do(newRequest)
					if err != nil {
						log.Println(err, resp)
					}
					if resp == nil {
						return
					}

					if resp.Header != nil {
						for k, v := range resp.Header {
							writer.Header().Add(k, strings.Join(v, ","))
						}
					}
					defer resp.Body.Close()
					io.Copy(writer, resp.Body)
					return
				}
			}
		}
	})

	http.Handle("/", g.Router)
	srv := &http.Server{
		Handler: g.Router,
		Addr:    "0.0.0.0:80",
	}
	go func() {
		if (g.certFile != "") && (g.keyFile != "") {
			srvTLS := &http.Server{
				Handler: g.Router,
				Addr:    "0.0.0.0:443",
			}
			fmt.Println("Starting HTTPS server on ", srv.Addr)
			srvTLS.ListenAndServeTLS(g.certFile, g.keyFile)
		}
	}()
	srv.ListenAndServe()
}
