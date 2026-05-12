package httpx

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type D map[string]string

func DecodeJSON[T any](r *http.Request, dst *T) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func sendJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json; charset=utf8")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func sendJSONError(w http.ResponseWriter, status int, err string) error {
	return sendJSON(w, status, D{
		"error": err,
	})
}

func sendText(w http.ResponseWriter, status int, text []byte) error {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, err := w.Write(text)
	return err
}

func sendHTML(w http.ResponseWriter, status int, html []byte) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, err := w.Write(html)
	return err
}

type Responder struct {
	logger *slog.Logger
}

func NewResponder(logger *slog.Logger) *Responder {
	if logger == nil {
		panic("logger is nil")
	}

	return &Responder{logger: logger}
}

func (res *Responder) JSON(w http.ResponseWriter, status int, data any) {
	if err := sendJSON(w, status, data); err != nil {
		res.logger.Error("sendJSON", "error", err)
	}
}

func (res *Responder) JSONError(w http.ResponseWriter, status int, errMsg string) {
	err := sendJSON(w, status, D{
		"error": errMsg,
	})
	if err != nil {
		res.logger.Error("sendJSONError", "error", err)
	}
}

func (res *Responder) Text(w http.ResponseWriter, status int, text []byte) {
	if err := sendText(w, status, text); err != nil {
		res.logger.Error("sendText", "error", err)
	}
}

func (res *Responder) HTML(w http.ResponseWriter, status int, html []byte) {
	if err := sendHTML(w, status, html); err != nil {
		res.logger.Error("sendHTML", "error", err)
	}
}
