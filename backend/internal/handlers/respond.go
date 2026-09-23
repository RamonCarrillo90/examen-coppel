package handlers

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

// writeJSON serializa data como JSON y lo escribe con el código HTTP indicado.
func writeJSON(w http.ResponseWriter, status int, data any) {
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

// writeError envía un error con el formato estándar de la API.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}
