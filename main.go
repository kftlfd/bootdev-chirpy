package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, err := w.Write([]byte("OK"))
		if err != nil {
			log.Println(err)
		}
	})

	cfg := &apiConfig{}

	mux.HandleFunc("GET /admin/metrics", cfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", cfg.resetMetricsHandler)

	mux.HandleFunc("POST /api/validate_chirp", func(w http.ResponseWriter, r *http.Request) {
		type reqBody struct {
			Body string `json:"body"`
		}
		decoder := json.NewDecoder(r.Body)
		body := reqBody{}
		if err := decoder.Decode(&body); err != nil {
			log.Printf("Error decoding parameters: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if len(body.Body) > 140 {
			type errRespoBody struct {
				Error string `json:"error"`
			}
			respBody := errRespoBody{Error: "Chirp is too long"}
			data, err := json.Marshal(respBody)
			if err != nil {
				log.Printf("Error marshalling JSON: %s", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json; charset=utf8")
			w.Write(data)
			return
		}

		type okResBody struct {
			Valid bool `json:"valid"`
		}
		respBody := okResBody{Valid: true}
		data, err := json.Marshal(respBody)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json; charset=utf8")
		w.Write(data)
	})

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
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileserverHits.Load())
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
