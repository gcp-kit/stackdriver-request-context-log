package main

import (
	"fmt"
	"net/http"

	log "github.com/gcp-kit/stalog"
)

func main() {
	mux := http.NewServeMux()

	projectId := "my-gcp-project"

	// Make config for this library
	config := log.NewConfig(projectId)
	config.AdditionalData = log.AdditionalData{ // set additional fields for all logs
		"service": "foo",
		"version": 1.0,
	}

	// Set request handler
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Get request context logger
		log.RequestLoggingWithFunc(config, w, r, index)
	})

	// Run server
	fmt.Println("Waiting requests on port 8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	logger := log.RequestContextLogger(r)

	// These logs are grouped with the request log
	logger.Debugf("Hi")
	logger.Infof("Hello")
	logger.Warnf("World")

	_, _ = fmt.Fprintf(w, "OK\n")
}
