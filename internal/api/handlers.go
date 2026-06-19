package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Adhnan23/karots-drop/internal/qr"
	"github.com/Adhnan23/karots-drop/internal/store"
)

const maxUploadSize = 20 << 20

type storeResponse struct {
	Code      string `json:"code"`
	TTL       int    `json:"ttl"`
	URL       string `json:"url"`
	Filename  string `json:"filename,omitempty"`
	Encrypted bool   `json:"encrypted"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorResponse{Error: msg})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleStore(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize+1<<20)

	ct := r.Header.Get("Content-Type")

	var (
		data        []byte
		filename    string
		contentType string
	)

	if strings.HasPrefix(ct, "multipart/form-data") {
		if err := r.ParseMultipartForm(maxUploadSize + 1<<20); err != nil {
			writeError(w, "Upload too large (max 20MB)", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err == nil {
			defer file.Close()
			data, err = io.ReadAll(file)
			if err != nil {
				writeError(w, "Failed to read file", http.StatusInternalServerError)
				return
			}
			filename = header.Filename
			if filename != "" {
				contentType = header.Header.Get("Content-Type")
			}
		}

		if len(data) == 0 {
			text := r.FormValue("text")
			if text == "" {
				writeError(w, "No text or file provided", http.StatusBadRequest)
				return
			}
			data = []byte(text)
			contentType = "text/plain; charset=utf-8"
		}
	} else {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, "Failed to read body", http.StatusInternalServerError)
			return
		}
		if len(body) == 0 {
			writeError(w, "Empty body", http.StatusBadRequest)
			return
		}
		data = body
		if ct != "" {
			contentType = ct
		} else {
			contentType = "application/octet-stream"
		}
	}

	if len(data) > maxUploadSize {
		writeError(w, "Data exceeds 20MB limit", http.StatusBadRequest)
		return
	}

	encrypted := r.FormValue("encrypted") == "true"

	code, err := s.store.GenerateUniqueCode()
	if err != nil {
		writeError(w, "Failed to generate code", http.StatusInternalServerError)
		return
	}

	itemTTL := s.ttl
	if ttlStr := r.FormValue("ttl"); ttlStr != "" {
		if d, err := time.ParseDuration(ttlStr); err == nil && d > 0 && d < itemTTL {
			itemTTL = d
		}
	}

	if err := s.store.SetWithTTL(code, data, filename, contentType, encrypted, itemTTL); err != nil {
		if errors.Is(err, store.ErrStoreFull) {
			writeError(w, "Store is full, try again later", http.StatusServiceUnavailable)
			return
		}
		writeError(w, "Failed to store data", http.StatusInternalServerError)
		return
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	url := fmt.Sprintf("%s://%s/api/get/%s", scheme, r.Host, code)

	resp := storeResponse{
		Code:      code,
		TTL:       int(itemTTL.Seconds()),
		URL:       url,
		Filename:  filename,
		Encrypted: encrypted,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		writeError(w, "Missing code", http.StatusBadRequest)
		return
	}

	item, ok := s.store.Get(code)
	if !ok {
		writeError(w, "Not found or expired", http.StatusNotFound)
		return
	}

	if s.deleteOnRetrieve {
		s.store.Delete(code)
	}

	if item.Encrypted {
		w.Header().Set("X-Encrypted", "true")
	}
	if item.ContentType != "" {
		w.Header().Set("Content-Type", item.ContentType)
	}
	if item.Filename != "" {
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, item.Filename))
	}

	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(item.Data)))
	w.Write(item.Data)
}

func (s *Server) handleQR(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		writeError(w, "Missing code", http.StatusBadRequest)
		return
	}

	_, ok := s.store.Get(code)
	if !ok {
		writeError(w, "Not found or expired", http.StatusNotFound)
		return
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	url := fmt.Sprintf("%s://%s/api/get/%s", scheme, r.Host, code)

	png, err := qr.PNG(url)
	if err != nil {
		writeError(w, "Failed to generate QR", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(png)))
	w.Write(png)
}
