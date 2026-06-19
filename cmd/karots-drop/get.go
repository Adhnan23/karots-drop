package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Adhnan23/karots-drop/internal/crypt"
)

func cmdGet(args []string) {
	var keyB64 string
	serverURL := "http://localhost:8080"
	jsonOut := false

	filtered := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--server" && i+1 < len(args):
			serverURL = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--server="):
			serverURL = args[i][len("--server="):]
		case args[i] == "--key" && i+1 < len(args):
			keyB64 = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--key="):
			keyB64 = args[i][len("--key="):]
		case args[i] == "--json":
			jsonOut = true
		default:
			filtered = append(filtered, args[i])
		}
	}

	if len(filtered) < 1 {
		log.Fatal("usage: karots-drop get <code>")
	}
	code := filtered[0]

	resp, err := http.Get(serverURL + "/api/get/" + code)
	if err != nil {
		log.Fatal("request:", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		log.Fatal("not found or expired")
	}
	if resp.StatusCode != http.StatusOK {
		log.Fatal("server error:", resp.StatusCode)
	}

	isEncrypted := resp.Header.Get("X-Encrypted") == "true"
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("read body:", err)
	}

	if jsonOut {
		info := struct {
			Code      string `json:"code"`
			Size      int    `json:"size"`
			Encrypted bool   `json:"encrypted"`
		}{
			Code:      code,
			Size:      len(body),
			Encrypted: isEncrypted,
		}
		json.NewEncoder(os.Stdout).Encode(info)
		return
	}

	if isEncrypted && keyB64 != "" {
		key, err := base64.StdEncoding.DecodeString(keyB64)
		if err != nil {
			log.Fatal("decode key:", err)
		}
		plaintext, err := crypt.Decrypt(body, key)
		if err != nil {
			log.Fatal("decrypt:", err)
		}
		os.Stdout.Write(plaintext)
		return
	}

	if isEncrypted {
		fmt.Fprintf(os.Stderr, "Data is encrypted. Use --key to decrypt.\n")
	}

	os.Stdout.Write(body)
}
