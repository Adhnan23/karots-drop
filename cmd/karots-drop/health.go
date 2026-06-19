package main

import (
	"flag"
	"net/http"
	"os"
)

func cmdHealth(args []string) {
	fs := flag.NewFlagSet("health", flag.ExitOnError)
	serverURL := fs.String("server", "http://localhost:8080", "server URL")
	fs.Parse(args)

	resp, err := http.Get(*serverURL + "/api/health")
	if err != nil {
		os.Exit(1)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		os.Exit(1)
	}
}
