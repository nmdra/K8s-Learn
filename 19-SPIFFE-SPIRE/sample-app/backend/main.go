package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

const (
	listenAddr = ":8443"
	socketPath = "unix:///run/spire/sockets/agent.sock"
)

func main() {
	ctx := context.Background()

	// Create an X509Source from the Workload API
	source, err := workloadapi.NewX509Source(ctx, workloadapi.WithClientOptions(workloadapi.WithAddr(socketPath)))
	if err != nil {
		log.Fatalf("Unable to create X509Source: %v", err)
	}
	defer source.Close()

	// Only allow connections from the frontend
	frontendID := spiffeid.RequireFromString("spiffe://example.org/ns/default/sa/frontend")
	tlsCfg := tlsconfig.MTLSServerConfig(source, source, tlsconfig.AuthorizeID(frontendID))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request from %s", r.RemoteAddr)
		fmt.Fprintf(w, "Hello from the backend! mTLS verified.\n")
	})

	server := &http.Server{
		Addr:      listenAddr,
		TLSConfig: tlsCfg,
	}

	log.Printf("Backend listening on %s with mTLS", listenAddr)
	// TLS is configured via the TLSConfig, so use ServeTLS with empty cert/key
	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
