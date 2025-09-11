package main

import (
	"crypto/tls"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type LogEntry struct {
	startTime time.Time
	endTime   time.Time
	note      string
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
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			io.WriteString(w, "===== Time Logger API =====\n")
			io.WriteString(w, "GET / - returns the Time Logger API instructions\n")
			io.WriteString(w, "GET /health - returns the service health\n")
			io.WriteString(w, "POST /log - adds a log entry, returns log id")
		} else {
			http.Error(w, "", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "", http.StatusMethodNotAllowed)
		}
		w.Write([]byte("Server is functioning optimally."))
	})

	mux.HandleFunc("/logs", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			if r.Header.Get("Content-Type") != "application/json" {
				http.Error(w, "", http.StatusUnsupportedMediaType)
			}

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
