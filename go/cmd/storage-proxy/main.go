package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mlaiops/platform/internal/storage"
)

func main() {
	config := storage.Config{Endpoint: os.Getenv("S3_ENDPOINT"), Region: env("S3_REGION", "us-east-1"), AccessKey: os.Getenv("S3_ACCESS_KEY"), SecretKey: os.Getenv("S3_SECRET_KEY")}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		write(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /presign", func(w http.ResponseWriter, r *http.Request) {
		if token := os.Getenv("MLAIOPS_INTERNAL_TOKEN"); token != "" && r.Header.Get("Authorization") != "Bearer "+token {
			write(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		var request storage.Request
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&request); err != nil {
			write(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		signed, err := storage.Presign(config, request, time.Now())
		if err != nil {
			write(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
			return
		}
		write(w, http.StatusOK, map[string]any{"url": signed, "expires_in": request.TTLSeconds})
	})
	server := &http.Server{Addr: ":" + env("PORT", "8084"), Handler: mux, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	log.Printf("storage proxy listening on %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}

func write(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
