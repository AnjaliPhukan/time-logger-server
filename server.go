package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var stdout *log.Logger = log.New(os.Stdout, "", log.Lmsgprefix)

type LogEntry struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Note      string    `json:"note"`
}

func rootFunc(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		io.WriteString(w, "===== Time Logger API =====\n")
		io.WriteString(w, "GET / - returns the Time Logger API instructions\n")
		io.WriteString(w, "GET /health - returns the service health\n")
		io.WriteString(w, "POST /log - adds a log entry, returns log id")
		return
	}
	http.Error(w, "", http.StatusMethodNotAllowed)
}

func logsFunc(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		if req.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "", http.StatusUnsupportedMediaType)
		}
		byteData, err := io.ReadAll(req.Body)
		if err != nil {
			log.Printf("Could not read request body: %v\n", err)
		}
		var entry LogEntry
		err = json.Unmarshal(byteData, &entry)
		if err != nil {
			log.Printf("Could not parse client JSON data as LogEntry struct: %v\n", err)
		}
		stdout.Println(entry)
		w.Write([]byte("Was able to read input json!"))

	default:
		http.Error(w, "", http.StatusMethodNotAllowed)
	}
}

func main() {
	port := flag.Int("port", 8443, "Port to open the server on.")
	certFile := flag.String("cert", "certs/server.crt", "Path to TLS certificate file.")
	keyFile := flag.String("key", "certs/server.key", "Path to server private key.")
	flag.Parse()

	cert, err := tls.LoadX509KeyPair(*certFile, *keyFile)
	if err != nil {
		stdout.Fatalf("Error while loading TLS key pair: %v\n", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", rootFunc)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "", http.StatusMethodNotAllowed)
		}
		w.Write([]byte("Server is functioning optimally."))
	})
	mux.HandleFunc("/logs", logsFunc)

	tlsConf := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
	}

	addr := ":" + strconv.Itoa(*port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		TLSConfig:    tlsConf,
	}
	stdout.Printf("Starting HTTP server at https://localhost:%s/", strconv.Itoa(*port))
	err = srv.ListenAndServeTLS("", "")
	if err != nil {
		stdout.Fatal("Error while starting server. Closing Program.\n")
	}
}
