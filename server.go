package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type LogEntry struct {
	startTime time.Time
	endTime   time.Time
}

func main() {
	port := flag.Int("port", 8443, "Port to open the server on.")
	certFile := flag.String("cert", "certs/server.crt", "Path to TLS certificate file.")
	keyFile := flag.String("key", "certs/server.key", "Path to server private key.")
	flag.Parse()

	addr := ":" + strconv.Itoa(*port)

	stdout := log.Logger{}
	stdout.SetOutput(os.Stdout)

	cert, err := tls.LoadX509KeyPair(*certFile, *keyFile)
	if err != nil {
		stdout.Fatalf("Error while loading TLS key pair: %v\n", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "", http.StatusMethodNotAllowed)
		} else {
			w.Write([]byte("Hello!!"))
		}
	})

	tlsConf := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
	}

	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		TLSConfig:    tlsConf,
	}
	stdout.Printf("Starting HTTP server at https://localhost%s/", addr)
	err = srv.ListenAndServeTLS("", "")
	if err != nil {
		stdout.Fatalln("Error while starting server. Closing Program.")
	}
}
