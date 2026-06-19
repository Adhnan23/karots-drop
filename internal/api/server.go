package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/Adhnan23/karots-drop/internal/store"
)

type Config struct {
	Addr             string
	Token            string
	RateLimit        int
	DeleteOnRetrieve bool
	TTL              time.Duration
}

type Server struct {
	*http.Server
	store            *store.Store
	token            string
	rateLimiter      *RateLimiter
	deleteOnRetrieve bool
	ttl              time.Duration
}

func New(cfg Config, s *store.Store, pageFS http.FileSystem, staticFS http.FileSystem) *Server {
	if cfg.TTL <= 0 {
		cfg.TTL = 20 * time.Minute
	}
	srv := &Server{
		store:            s,
		token:            cfg.Token,
		deleteOnRetrieve: cfg.DeleteOnRetrieve,
		ttl:              cfg.TTL,
	}

	if cfg.RateLimit > 0 {
		srv.rateLimiter = NewRateLimiter(cfg.RateLimit, 1*time.Minute)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/store", srv.handleStore)
	mux.HandleFunc("GET /api/get/{code}", srv.handleGet)
	mux.HandleFunc("GET /api/qr/{code}", srv.handleQR)
	mux.HandleFunc("GET /api/health", srv.handleHealth)

	if pageFS != nil {
		mux.HandleFunc("GET /", servePage(pageFS, "index.html"))
		mux.HandleFunc("GET /file", servePage(pageFS, "file.html"))
		mux.HandleFunc("GET /retrieve", servePage(pageFS, "retrieve.html"))
		mux.HandleFunc("GET /docs", servePage(pageFS, "docs.html"))
	}
	if staticFS != nil {
		mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(staticFS)))
	}

	var handler http.Handler = mux
	handler = srv.rateLimitMiddleware(handler)
	handler = srv.authMiddleware(handler)
	handler = corsMiddleware(handler)
	handler = loggingMiddleware(handler)

	srv.Server = &http.Server{
		Addr:         cfg.Addr,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return srv
}

func servePage(fsys http.FileSystem, name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && r.URL.Path != "/file" && r.URL.Path != "/retrieve" && r.URL.Path != "/docs" {
			http.NotFound(w, r)
			return
		}
		f, err := fsys.Open(name)
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		defer f.Close()
		stat, _ := f.Stat()
		http.ServeContent(w, r, name, stat.ModTime(), f)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Auth-Token")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.token == "" {
			next.ServeHTTP(w, r)
			return
		}
		if !strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/api/health" {
			next.ServeHTTP(w, r)
			return
		}
		token := r.Header.Get("X-Auth-Token")
		if token != s.token {
			writeError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
