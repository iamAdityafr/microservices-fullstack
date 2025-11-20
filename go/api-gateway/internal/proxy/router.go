package proxy

import (
	"api-gateway/internal/config"
	"api-gateway/internal/middleware"
	"context"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type Router struct {
	cfg *config.Config
}

func NewRouter(cfg *config.Config) http.Handler {
	r := &Router{cfg: cfg}
	mux := http.NewServeMux()

	authMW := middleware.AuthMiddleware(cfg.JWTSecret, nil)

	// Public routes
	mux.HandleFunc("/register", r.handleProxy(cfg.UserServiceURL))
	mux.HandleFunc("/login", r.handleProxy(cfg.UserServiceURL))

	// Products - public rn
	mux.HandleFunc("/products/get", r.handleProxy(cfg.ProductServiceURL))
	mux.HandleFunc("/products/search", r.handleProxy(cfg.ProductServiceURL))

	// Protected routes
	protectedRoutes := map[string]string{
		"/profile":          cfg.UserServiceURL,
		"/users/":           cfg.UserServiceURL,
		"/orders/":          cfg.OrderServiceURL,
		"/orders":           cfg.OrderServiceURL,
		"/cart/getcart":     cfg.CartServiceURL,
		"/cart/add":         cfg.CartServiceURL,
		"/cart":             cfg.CartServiceURL,
		"/cart/":            cfg.CartServiceURL,
		"/payments/":        cfg.PaymentServiceURL,
		"/payments/intent":  cfg.PaymentServiceURL,
		"/payments/webhook": cfg.PaymentServiceURL,
		"/payments":         cfg.PaymentServiceURL,
		"/notifications/":   cfg.NotificationServiceURL,
		"/notifications":    cfg.NotificationServiceURL,
	}

	for route, serviceURL := range protectedRoutes {
		handler := r.handleProxy(serviceURL)
		mux.Handle(route, authMW(http.HandlerFunc(handler)))
	}

	return mux
}

func (r *Router) handleProxy(targetURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		log.Printf("Proxy handle request: %s %s", req.Method, req.URL.Path)
		err := r.forwardRequest(w, req, targetURL)
		if err != nil {
			log.Printf("proxy err: %v", err)
			http.Error(w, "service err", http.StatusServiceUnavailable)
		}
	}
}

func (r *Router) forwardRequest(w http.ResponseWriter, req *http.Request, targetHost string) error {
	targetHost = strings.TrimRight(targetHost, "/")

	path := req.URL.Path
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	url := targetHost + path
	if req.URL.RawQuery != "" {
		url += "?" + req.URL.RawQuery
	}

	log.Printf("[forward]: %s %s -> %s", req.Method, req.URL.Path, url)
	ctx, cancel := context.WithTimeout(req.Context(), 30*time.Second)
	defer cancel()

	pReq, err := http.NewRequestWithContext(ctx, req.Method, url, req.Body)
	if err != nil {
		log.Printf("[Error] creating proxy request: %v", err)
		return err
	}

	// copuing headers
	for k, vals := range req.Header {
		for _, v := range vals {
			pReq.Header.Add(k, v)
		}
	}

	pReq.Header.Set("X-Forwarded-Proto", "http")
	pReq.Header.Set("X-Forwarded-Host", req.Host)
	pReq.Header.Set("X-Forwarded-For", req.RemoteAddr)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(pReq)
	if err != nil {
		log.Printf("[Error]: executing proxy request: %v", err)
		return err
	}
	defer resp.Body.Close()

	log.Printf("[got resp]: %d from %s", resp.StatusCode, url)

	for k, vals := range resp.Header {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	return err
}
