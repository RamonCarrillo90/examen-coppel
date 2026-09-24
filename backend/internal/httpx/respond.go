// Package httpx contiene ayudantes para responder y leer JSON en HTTP
package httpx

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// errorResponse es el formato único de error que devuelve toda la API.
// Tener un solo formato facilita que el frontend muestre los mensajes.
//
//	{ "error": "el email ya está registrado" }
type errorResponse struct {
	Error string `json:"error"`
}

// WriteJSON serializa data como JSON y lo escribe con el código HTTP indicado.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if data == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// En este punto ya se enviaron los encabezados; solo queda registrarlo.
		slog.Error("no se pudo escribir la respuesta JSON", "error", err)
	}
}

// DecodeJSON lee el cuerpo de la petición como JSON en dst, con límites de seguridad.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // máximo 1 MB
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields() // rechaza campos que no existen en el DTO
	return dec.Decode(dst)
}

// WriteError envía un error con el formato estándar de la API.
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, errorResponse{Error: message})
}
