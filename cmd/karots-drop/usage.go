package main

import (
	"fmt"
	"os"
)

func printUsage() {
	fmt.Fprintf(os.Stderr, `karots-drop %s - ephemeral text & file sharing

Usage:
  karots-drop serve [flags]
  karots-drop send [flags] [text]
  karots-drop get <code> [flags]
  karots-drop clip [flags]
  karots-drop version
  karots-drop health
  karots-drop completion

Commands:
  serve       Start the HTTP server with API and Web UI
  send        Upload text, a file, or piped data
  get         Download data by 6-digit code
  clip        Read clipboard and upload
  version     Print version
  health      Check server health
  completion  Print bash completion script

Flags:
  Use --help with any command for command-specific flags.
  See README.md or https://github.com/Adhnan23/karots-drop for full docs.
`, version)
}
