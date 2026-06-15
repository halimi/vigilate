package main

import (
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

const (
	vigilateVersion = "0.1.0"
)

func init() {
	_ = os.Setenv("TZ", "Europe/Amsterdam")
}

func main() {
	insecurePort, err := setupApp()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("**************************************************")
	log.Printf("** %sVigilate%s v%s built in %s **", "\033[31m", "\033[0m", vigilateVersion, runtime.Version())
	log.Printf("**----------------------------------------------**")
	log.Printf("** Running with %d Processors                   **", runtime.NumCPU())
	log.Printf("** Running on %s                             **", runtime.GOOS)
	log.Printf("**************************************************")

	srv := &http.Server{
		Addr:              *insecurePort,
		Handler:           routes(),
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	log.Printf("Strating HTTP server on port %s", *insecurePort)

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
