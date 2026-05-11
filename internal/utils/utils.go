package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

type D map[string]string

func SendJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf8")

	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("json marshall: %v", err)
	}
}

func SendJSONError(w http.ResponseWriter, status int, err string) {
	SendJSON(w, status, D{
		"error": err,
	})
}

func DecodeJSON[T any](r *http.Request, dst *T) error {
	dec := json.NewDecoder(r.Body)

	dec.DisallowUnknownFields()

	return dec.Decode(dst)
}
