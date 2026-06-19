package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/Adhnan23/karots-drop/internal/crypt"
	"github.com/Adhnan23/karots-drop/internal/qr"
)

func cmdSend(args []string) {
	fs := flag.NewFlagSet("send", flag.ExitOnError)
	filePath := fs.String("file", "", "path to file")
	serverURL := fs.String("server", "http://localhost:8080", "server URL")
	encrypt := fs.Bool("encrypt", false, "encrypt data before upload")
	showQR := fs.Bool("qr", false, "show QR code")
	jsonOut := fs.Bool("json", false, "output as JSON")
	compact := fs.Bool("compact", false, "output only the code")
	ttlStr := fs.String("ttl", "", "request custom TTL (e.g. 5m, 1h)")
	fs.Parse(args)

	var (
		data     []byte
		filename string
		isFile   bool
	)

	if *filePath != "" {
		var err error
		data, err = os.ReadFile(*filePath)
		if err != nil {
			log.Fatal("read file:", err)
		}
		filename = *filePath
		if idx := strings.LastIndex(filename, "/"); idx >= 0 {
			filename = filename[idx+1:]
		}
		isFile = true
	} else if fs.NArg() > 0 {
		data = []byte(strings.Join(fs.Args(), " "))
	} else if isStdinPipe() {
		var err error
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal("read stdin:", err)
		}
	} else {
		log.Fatal("no input: provide text, --file, or pipe data")
	}

	var encKey []byte
	if *encrypt {
		key, err := crypt.GenerateKey()
		if err != nil {
			log.Fatal("generate key:", err)
		}
		encKey = key
		ciphertext, err := crypt.Encrypt(data, key)
		if err != nil {
			log.Fatal("encrypt:", err)
		}
		data = ciphertext
	}

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	if isFile {
		part, err := w.CreateFormFile("file", filename)
		if err != nil {
			log.Fatal("multipart:", err)
		}
		part.Write(data)
	} else {
		w.WriteField("text", string(data))
	}

	if *encrypt {
		w.WriteField("encrypted", "true")
	}
	if *ttlStr != "" {
		w.WriteField("ttl", *ttlStr)
	}
	w.Close()

	req, err := http.NewRequest("POST", *serverURL+"/api/store", &buf)
	if err != nil {
		log.Fatal("create request:", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("upload:", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		log.Fatalf("upload failed (%d): %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var result struct {
		Code      string `json:"code"`
		TTL       int    `json:"ttl"`
		URL       string `json:"url"`
		Filename  string `json:"filename,omitempty"`
		Encrypted bool   `json:"encrypted"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatal("decode response:", err)
	}

	if *compact {
		fmt.Print(result.Code)
		return
	}

	if *jsonOut {
		output := struct {
			Code      string `json:"code"`
			TTL       int    `json:"ttl"`
			URL       string `json:"url"`
			Filename  string `json:"filename,omitempty"`
			Encrypted bool   `json:"encrypted"`
			Key       string `json:"key,omitempty"`
		}{
			Code:      result.Code,
			TTL:       result.TTL,
			URL:       result.URL,
			Filename:  result.Filename,
			Encrypted: result.Encrypted,
		}
		if *encrypt && encKey != nil {
			output.Key = base64.StdEncoding.EncodeToString(encKey)
		}
		json.NewEncoder(os.Stdout).Encode(output)
		return
	}

	fmt.Printf("Code: %s\n", result.Code)
	fmt.Printf("URL:  %s\n", result.URL)
	fmt.Printf("TTL:  %d seconds\n", result.TTL)
	if result.Filename != "" {
		fmt.Printf("File: %s\n", result.Filename)
	}
	if *encrypt && encKey != nil {
		keyB64 := base64.StdEncoding.EncodeToString(encKey)
		fmt.Printf("Key:  %s\n", keyB64)
		fmt.Println("Decrypt with: karots-drop get " + result.Code + " --key " + keyB64)
	}

	if *showQR {
		qrStr, err := qr.Terminal(result.URL)
		if err != nil {
			log.Fatal("qr:", err)
		}
		fmt.Println("\nQR Code:")
		fmt.Println(qrStr)
	}
}

func isStdinPipe() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) == 0
}
