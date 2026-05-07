package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, err := w.Write([]byte("OK"))
		if err != nil {
			log.Println(err)
		}
	})

	cfg := &apiConfig{}

	mux.HandleFunc("/metrics", cfg.metricsHandler)
	mux.HandleFunc("/reset", cfg.resetMetricsHandler)

	mux.Handle("/app/", cfg.middlewareMetricsInc(newAppFsHandler("/app/")))

	server := http.Server{Handler: mux, Addr: ":8080"}
	log.Fatal(server.ListenAndServe())
}

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	data := fmt.Sprintf("Hits: %d\n", cfg.fileserverHits.Load())
	w.Write([]byte(data))
}

func (cfg *apiConfig) resetMetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
}

func newAppFsHandler(prefix string) http.Handler {
	return http.StripPrefix(prefix, http.FileServer(http.Dir(".")))
}

func middlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
