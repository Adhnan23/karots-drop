package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Adhnan23/karots-drop/internal/api"
	"github.com/Adhnan23/karots-drop/internal/store"
	"github.com/Adhnan23/karots-drop/internal/webui"
)

func cmdServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	addr := fs.String("addr", ":8080", "listen address")
	token := fs.String("token", "", "API token for authentication (empty = disabled)")
	rateLimit := fs.Int("rate-limit", 60, "max requests per minute per IP (0 = unlimited)")
	deleteOnRetrieve := fs.Bool("delete-on-retrieve", false, "delete items after retrieval")
	maxItems := fs.Int("max-items", 500, "max items in store (0 = unlimited)")
	ttl := fs.Duration("ttl", 20*time.Minute, "item TTL (e.g. 5m, 30m, 1h)")
	bind := fs.String("bind", "", "bind to interface (e.g. localhost)")
	fs.Parse(args)

	listenAddr := *addr
	if strings.ToLower(*bind) == "localhost" {
		listenAddr = "127.0.0.1" + listenAddr
	}

	var s *store.Store
	if *maxItems > 0 {
		s = store.NewWithMaxAndTTL(*maxItems, *ttl)
	} else {
		s = store.NewWithTTL(*ttl)
	}
	defer s.Stop()

	srv := api.New(api.Config{
		Addr:             listenAddr,
		Token:            *token,
		RateLimit:        *rateLimit,
		DeleteOnRetrieve: *deleteOnRetrieve,
		TTL:              *ttl,
	}, s, webui.FileSystem(), webui.StaticFS())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		fmt.Fprintln(os.Stderr, "\nshutting down...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Fatal("shutdown:", err)
		}
	}()

	fmt.Fprintf(os.Stderr, "karots-drop listening on %s\n", listenAddr)
	fmt.Fprintf(os.Stderr, "Web UI: http://localhost%s\n", listenAddr)
	fmt.Fprintf(os.Stderr, "API:    http://localhost%s/api/\n", listenAddr)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
