package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/cbergoon/glb/proxy"
	"github.com/cbergoon/glb/registry"
)

// ServiceRegistry is a local registry of services/versions
var ServiceRegistry = registry.DefaultRegistry{
	"service1": {
		"v1": {
			"localhost:9091",
			"localhost:9092",
		},
	},
}

func runLoadBalancer() {
	go http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req,
			"https://localhost:8443"+req.URL.String(),
			http.StatusMovedPermanently)
	}))

	http.HandleFunc("/", proxy.NewMultipleHostReverseProxy(ServiceRegistry))
	http.HandleFunc("/health", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "%v\n", ServiceRegistry)
	})

	log.Fatal(http.ListenAndServeTLS(":8443", "server.crt", "server.key", nil))
}

func main() {
	runLoadBalancer()
}
