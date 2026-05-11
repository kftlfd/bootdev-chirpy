package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

func SendJSON(w http.ResponseWriter, status int, data any) {
	body, err := json.Marshal(data)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json; charset=utf8")
	w.Write(body)
}
