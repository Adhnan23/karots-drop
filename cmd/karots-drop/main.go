package main

import (
	"fmt"
	"os"
)

var version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	subcommand := os.Args[1]
	args := os.Args[2:]

	switch subcommand {
	case "serve":
		cmdServe(args)
	case "send":
		cmdSend(args)
	case "get":
		cmdGet(args)
	case "clip":
		cmdClip(args)
	case "version", "--version", "-v":
		fmt.Println("karots-drop " + version)
	case "health":
		cmdHealth(args)
	case "completion":
		cmdCompletion(args)
	default:
		printUsage()
		os.Exit(1)
	}
}
