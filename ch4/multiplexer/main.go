package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var (
	hostProxy = make(map[string]string)
	proxies   = make(map[string]*httputil.ReverseProxy)
)

func init() {
	// in that example we are using 127.0.0.1 for both but you can modify if needed
	hostProxy["attacker1.com"] = "http://127.0.0.1:10080"
	hostProxy["attacker2.com"] = "http://127.0.0.1:20080"

	for k, v := range hostProxy {
		remote, err := url.Parse(v)
		if err != nil {
			log.Fatal("Unable to parse proxy target")
		}
		proxies[k] = httputil.NewSingleHostReverseProxy(remote)
	}
}

func main() {
	r := mux.NewRouter()
	for host, proxy := range proxies {
		r.Host(host).Handler(proxy)
	}
	log.Fatal(http.ListenAndServe("172.28.128.1:80", r))
}

// TODO: Use staged payload, This likely comes with additional challenges, as you’ll need to ensure that both stages are properly routed through the proxy
// TODO: Implement it by using HTTPS instead of cleartext HTTP