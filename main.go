//   Copyright 2019 The SlugBus++ Authors.

//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at

//        http://www.apache.org/licenses/LICENSE-2.0

//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/slugbus/taps"
)

// Secure setup thanks to https://blog.cloudflare.com/exposing-go-on-the-internet/
func setupServer(port uint64) *http.Server {

	tlsConfig := &tls.Config{
		// Causes servers to use Go's default ciphersuite preferences,
		// which are tuned to avoid attacks. Does nothing on clients.
		PreferServerCipherSuites: true,
		// Only use curves which have assembly implementations
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
			tls.X25519, // Go 1.8 only
		},
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305, // Go 1.8 only
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,   // Go 1.8 only
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,

			// Best disabled, as they don't provide Forward Secrecy,
			// but might be necessary for some clients
			// tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			// tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		},
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%d", port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		TLSConfig:    tlsConfig,
	}

	return srv
}

func main() {
	// Define the flags
	port := flag.Uint64("port", 8080, "port to listen on")
	dataFile := flag.String("data", "data/feb-25-interval-3s-1850-queries.json", "file to use as mock data")
	duration := flag.Duration("interval", 3*time.Second, "the interval that the mock data is spaced apart")
	flag.Parse()

	// Read in the data
	mockBytes, err := ioutil.ReadFile(*dataFile)
	if err != nil {
		log.Fatal(err)
	}

	// Unmarshal the bytes into a struct
	mockResponses := [][]taps.Bus{}
	if err := json.Unmarshal(mockBytes, &mockResponses); err != nil {
		log.Fatal(err)
	}

	// Keep track of the "current state"
	currentIndex := 0
	// Update it asynchronously, using a muxtex to prevent
	// later data races
	mutex := &sync.Mutex{}
	go func() {
		for range time.Tick(*duration) {
			// Lock the mutex
			mutex.Lock()
			currentIndex = (currentIndex + 1 + len(mockResponses)) % len(mockResponses)
			mutex.Unlock()
		}
	}()

	// Setup a server
	srv := setupServer(*port)

	// Setup a mux
	mux := http.NewServeMux()

	// Use closures to response to http requests
	mux.HandleFunc("/location/get", func(w http.ResponseWriter, r *http.Request) {
		// Write the header
		w.Header().Set("Content-Type", "application/json")
		// Marshall the data
		mutex.Lock()
		responseBytes, err := json.Marshal(mockResponses[currentIndex])
		mutex.Unlock()
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := w.Write([]byte("[]")); err != nil {
				log.Println(err)
			}
			return
		}
		if _, err := w.Write(responseBytes); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}
	})

	// Add the mux to the server
	srv.Handler = mux

	// Start the server
	log.Printf("Using file %s as mock data\n", *dataFile)
	log.Printf("Data points are updated ~every %v\n", *duration)
	log.Printf("Starting server on %s\n", srv.Addr)
	log.Printf("Send queries to http://%s/location/get\n", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
