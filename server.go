package main

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
)

type argonParams struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

type LogEntry struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Note      string    `json:"note"`
}

var (
	stdout *log.Logger  = log.New(os.Stdout, "", log.Lmsgprefix)
	stderr *log.Logger  = log.New(os.Stderr, "", log.LstdFlags)
	params *argonParams = &argonParams{
		memory:      64 * 1024,
		iterations:  3,
		parallelism: 1,
		saltLength:  16,
		keyLength:   32,
	}
	EncodedHashParseError   = errors.New("Could not parse parameters and hash from encoded string.")
	IncompatibleArgonVerion = errors.New("Argon version used in hash is not compatible with current program's argon version.")
)

func decodeHash(encodedHash string) (p *argonParams, salt []byte, hash []byte, err error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, nil, EncodedHashParseError
	}

	if vals[1] != "argon2id" {
		return nil, nil, nil, EncodedHashParseError
	}

	var version int
	n, err := fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	} else if n != 1 {
		return nil, nil, nil, EncodedHashParseError
	}
	if version != argon2.Version {
		return nil, nil, nil, IncompatibleArgonVerion
	}

	p = &argonParams{}
	n, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism)
	if err != nil {
		return nil, nil, nil, err
	} else if n != 3 {
		return nil, nil, nil, EncodedHashParseError
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, err
	}
	p.saltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, err
	}
	p.keyLength = uint32(len(hash))

	return p, salt, hash, nil
}

func rootHandler(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		errorPage, err := os.ReadFile("public/error.html")
		if err != nil {
			stderr.Printf("%v\n", err)
			http.Error(w, "404 Error - Page not found", http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(errorPage)
		}
	}
	switch req.Method {
	case http.MethodGet:
		landing, err := os.ReadFile("public/landing.html")
		if err != nil {
			stdout.Println("Could not find landing.html page.")
		} else {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(landing)
		}
	case http.MethodPost:
		err := req.ParseForm()
		if err != nil {
			stdout.Printf("%v\n", err)
		}
		username := req.PostForm.Get("username")
		password := req.PostForm.Get("password")
		if (username == "") || (password == "") {
			stdout.Print("The password and/or username could not be parsed from the input.")
		}
		saltBytes := make([]byte, params.saltLength)
		n, _ := rand.Read(saltBytes)
		if n != int(params.saltLength) {
		}
		hash := argon2.IDKey([]byte(password), saltBytes, params.iterations, params.memory, params.parallelism, params.keyLength)
		b64salt := base64.RawStdEncoding.EncodeToString(saltBytes)
		b64hash := base64.RawStdEncoding.EncodeToString(hash)
		encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, params.memory, params.iterations, params.parallelism, b64salt, b64hash)
		stdout.Print(encodedHash)
	default:
		http.Error(w, "", http.StatusMethodNotAllowed)
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		register, err := os.ReadFile("public/register.html")
		if err != nil {
			stdout.Println("Could not find the register.html page.")
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(register)
		}
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
	fileServer := http.FileServer(http.Dir("public/"))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "", http.StatusMethodNotAllowed)
		}
		w.Write([]byte("Server is functioning optimally."))
	})
	mux.HandleFunc("/register", registerHandler)

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
	stdout.Printf("Starting HTTP server at https://localhost%s/", addr)
	err = srv.ListenAndServeTLS("", "")
	if err != nil {
		stdout.Fatal("Error while starting server. Closing Program.\n")
	}
}
