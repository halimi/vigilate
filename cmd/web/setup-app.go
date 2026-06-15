package main

import "flag"

func setupApp() (*string, error) {
	insecurePort := flag.String("port", ":4000", "port to listen on")

	flag.Parse()

	return insecurePort, nil
}
