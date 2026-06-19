package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Adhnan23/karots-drop/internal/clip"
)

func cmdClip(args []string) {
	fs := flag.NewFlagSet("clip", flag.ExitOnError)
	serverURL := fs.String("server", "http://localhost:8080", "server URL")
	encrypt := fs.Bool("encrypt", false, "encrypt data before upload")
	showQR := fs.Bool("qr", false, "show QR code")
	jsonOut := fs.Bool("json", false, "output as JSON")
	watch := fs.Bool("watch", false, "watch clipboard for changes and auto-upload")
	fs.Parse(args)

	if *watch {
		clipWatch(*serverURL, *encrypt, *showQR, *jsonOut)
		return
	}

	text, err := clip.Read()
	if err != nil {
		log.Fatal(err)
	}
	if text == "" {
		log.Fatal("clipboard is empty")
	}

	sendArgs := []string{"--server", *serverURL}
	if *encrypt {
		sendArgs = append(sendArgs, "--encrypt")
	}
	if *showQR {
		sendArgs = append(sendArgs, "--qr")
	}
	if *jsonOut {
		sendArgs = append(sendArgs, "--json")
	}
	sendArgs = append(sendArgs, text)

	cmdSend(sendArgs)
}

func clipWatch(serverURL string, encrypt, showQR, jsonOut bool) {
	var last string
	fmt.Fprintln(os.Stderr, "watching clipboard... (press Ctrl+C to stop)")
	for {
		time.Sleep(2 * time.Second)
		text, err := clip.Read()
		if err != nil || text == "" || text == last {
			continue
		}
		last = text

		args := []string{"--server", serverURL}
		if encrypt {
			args = append(args, "--encrypt")
		}
		if showQR {
			args = append(args, "--qr")
		}
		if jsonOut {
			args = append(args, "--json")
		}
		args = append(args, text)

		fmt.Fprintf(os.Stderr, "\nuploading: %.60s\n", text)
		cmdSend(args)
	}
}
