package main

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fanscontest/qaragon-mcp/internal/catalog"
	"github.com/fanscontest/qaragon-mcp/internal/mcpserver"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	defaultListenAddr = ":8080"
	defaultContract   = "/app/generated/public-openapi.json"
	defaultEvents     = "/app/generated/webhook-events.json"
	defaultGuides     = "/app/content/guides"
	defaultSwaggerUI  = "/app/content/swagger-ui"
	maxRequestBytes   = 1 << 20
)

//go:embed apidocs.html
var apiDocsHTML []byte

func main() {
	if err := run(); err != nil {
		slog.Error("qaragon-mcp stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	api, err := catalog.Load(envOr("OPENAPI_PATH", defaultContract), envOr("WEBHOOK_EVENTS_PATH", defaultEvents))
	if err != nil {
		return fmt.Errorf("load public catalog: %w", err)
	}
	apiURL := envOr("PLATFORM_API_URL", "localhost")
	apiScheme := "https"
	if apiURL == "localhost" || apiURL == "127.0.0.1" {
		apiScheme = "http"
	}
	api.Document["servers"] = []any{map[string]any{"url": apiScheme + "://" + apiURL}}
	openAPIDocument, err := api.DocumentJSON()
	if err != nil {
		return fmt.Errorf("format public API contract: %w", err)
	}

	srv := mcpserver.New(api, envOr("GUIDES_DIR", defaultGuides))
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return srv
	}, &mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true})

	allowedHosts := splitSet(envOr("MCP_ALLOWED_HOSTS", "localhost,127.0.0.1,mcp.qaragon.com"))
	limiter := newIPLimiter(120, time.Minute, 4096)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=300")
		_, _ = w.Write(apiDocsHTML)
	})
	mux.Handle("GET /docs-assets/", http.StripPrefix("/docs-assets/", http.FileServer(http.Dir(defaultSwaggerUI))))
	mux.HandleFunc("GET /openapi.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.oai.openapi+json;version=3.0")
		w.Header().Set("Cache-Control", "public, max-age=300")
		_, _ = w.Write(openAPIDocument)
	})
	mux.Handle("/mcp", maxBytesHandler(maxRequestBytes, handler))

	root := hostGuard(allowedHosts, cors(limiter.middleware(mux)))
	server := &http.Server{
		Addr:              envOr("HTTP_ADDR", defaultListenAddr),
		Handler:           root,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       20 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    32 << 10,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		slog.Info("Qaragon public MCP listening", "addr", server.Addr, "operations", api.OperationCount())
		errCh <- server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shut down HTTP server: %w", err)
		}
		return nil
	}
}

func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func splitSet(value string) map[string]struct{} {
	values := strings.Split(value, ",")
	set := make(map[string]struct{}, len(values))
	for _, item := range values {
		item = strings.ToLower(strings.TrimSpace(item))
		if item != "" {
			set[item] = struct{}{}
		}
	}
	return set
}

func maxBytesHandler(limit int64, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			r.Body = http.MaxBytesReader(w, r.Body, limit)
		}
		next.ServeHTTP(w, r)
	})
}

func hostGuard(allowed map[string]struct{}, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := strings.ToLower(strings.TrimSpace(r.Host))
		if parsed, _, err := net.SplitHostPort(host); err == nil {
			host = parsed
		} else if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
			host = strings.TrimSuffix(strings.TrimPrefix(host, "["), "]")
		}
		if _, ok := allowed[host]; !ok {
			http.Error(w, "invalid host", http.StatusMisdirectedRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type ipLimiter struct {
	mu         sync.Mutex
	limit      int
	window     time.Duration
	maxEntries int
	clients    map[string]ipBucket
}

type ipBucket struct {
	started time.Time
	count   int
}

func newIPLimiter(limit int, window time.Duration, maxEntries int) *ipLimiter {
	return &ipLimiter{
		limit:      limit,
		window:     window,
		maxEntries: maxEntries,
		clients:    make(map[string]ipBucket),
	}
}

func (l *ipLimiter) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		now := time.Now()
		l.mu.Lock()
		bucket, exists := l.clients[ip]
		if !exists || now.Sub(bucket.started) >= l.window {
			bucket = ipBucket{started: now}
		}
		bucket.count++
		allowed := bucket.count <= l.limit
		l.clients[ip] = bucket
		if len(l.clients) > l.maxEntries {
			for address, entry := range l.clients {
				if now.Sub(entry.started) >= l.window {
					delete(l.clients, address)
				}
			}
			if len(l.clients) > l.maxEntries {
				// Drop the oldest window to keep memory bounded during source churn.
				var oldestIP string
				var oldest time.Time
				for address, entry := range l.clients {
					if oldestIP == "" || entry.started.Before(oldest) {
						oldestIP, oldest = address, entry.started
					}
				}
				delete(l.clients, oldestIP)
			}
		}
		l.mu.Unlock()

		if !allowed {
			w.Header().Set("Retry-After", strconv.Itoa(max(1, int(l.window.Seconds()))))
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	// The shared edge overwrites X-Real-IP before proxying. Fall back to the
	// socket address for local use and direct health checks.
	if forwarded := strings.TrimSpace(r.Header.Get("X-Real-IP")); forwarded != "" {
		if ip := net.ParseIP(forwarded); ip != nil {
			return ip.String()
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Mcp-Protocol-Version, Last-Event-ID")
		w.Header().Set("Access-Control-Expose-Headers", "Mcp-Protocol-Version")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
