package main

import (
	"fmt"
	"log"
	"net/http"

	"os"

	"github.com/cbergoon/glb/proxy"
	"github.com/cbergoon/glb/registry"
)

const (
	CONFIG_FILE = "glb.json" //File containing configurataion
	CERT_FILE = "server.cert" //SSL Certificate
	KEY_FILE = "server.key" //SSL Key
)

var ServiceRegistry *registry.DefaultRegistry = &registry.DefaultRegistry{} //Service registry to store service-address mappings.
var BasicProxy bool = false //Enable single service "default" service/version. Removes requirement of service/version in URL.
var IdleConnTimeoutSeconds int = 1 //Duration the transport should keep connections alive. Zero imposes no limit.
var DisableKeepAlives bool = false //Do not keep alive, reconnect on each request.

//Starts load balancer, redirect for HTTPS and, service endpoints.
func runLoadBalancer(addr, port, sslPort string) {
	//Redirect to HTTPS
	if sslPort != "" {
		go http.ListenAndServe(port, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			http.Redirect(w, req,
				"https://"+addr+sslPort+req.URL.String(),
				http.StatusMovedPermanently)
		}))
	}
	//GLB Service Endpoints
	http.HandleFunc("/status", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "%v\n", *ServiceRegistry)
	})
	http.HandleFunc("/reload", func(w http.ResponseWriter, req *http.Request) {
		config, err := ReadParseConfig(CONFIG_FILE)
		if err != nil {
			log.Print(err)
			os.Exit(-1)
		}
		*ServiceRegistry = config.Registry
		BasicProxy = config.Basic
		IdleConnTimeoutSeconds = config.IdleConnTimeoutSeconds
		DisableKeepAlives = config.DisableKeepAlives
		fmt.Fprintf(w, "%v\n", *ServiceRegistry)
	})
	//Proxy Endpoint
	http.HandleFunc("/", proxy.NewMultipleHostReverseProxy(ServiceRegistry, &BasicProxy, &IdleConnTimeoutSeconds, &DisableKeepAlives))
	if sslPort != "" {
		log.Fatal(http.ListenAndServeTLS(sslPort, CERT_FILE, KEY_FILE, nil))
	}else{
		log.Fatal(http.ListenAndServe(sslPort, nil))
	}
}

func main() {
	//Configure
	config, err := ReadParseConfig(CONFIG_FILE)
	if err != nil {
		log.Print(err)
		os.Exit(-1)
	}
	*ServiceRegistry = config.Registry
	BasicProxy = config.Basic
	IdleConnTimeoutSeconds = config.IdleConnTimeoutSeconds
	DisableKeepAlives = config.DisableKeepAlives
	//Run
	runLoadBalancer(config.Host.Addr, config.Host.Port, config.Host.SslPort)
}
