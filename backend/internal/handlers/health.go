// Package handlers contiene los manejadores HTTP de la API: reciben la
// petición, validan la entrada, llaman a la capa de datos y responden JSON.
package handlers

import (
	"context"
	"net/http"
	"time"
)

// Pinger es cualquier cosa capaz de verificar que la base de datos responde.
//
// Usamos una interfaz (y no *pgxpool.Pool directamente) por dos razones:
//  1. El handler no depende de una librería concreta de base de datos.
//  2. En las pruebas podemos pasar un Pinger falso sin levantar PostgreSQL.
//
// *pgxpool.Pool ya tiene el método Ping(ctx) error, así que cumple la interfaz
// sin escribir nada extra.
type Pinger interface {
	Ping(ctx context.Context) error
}

// HealthHandler responde si la API y la base de datos están funcionando.
// Docker, Nginx o un balanceador pueden consultarlo para saber si el servicio
// está sano.
type HealthHandler struct {
	db Pinger
}

// NewHealthHandler crea el handler. db puede ser nil mientras la conexión a la
// base de datos todavía no esté implementada.
func NewHealthHandler(db Pinger) *HealthHandler {
	return &HealthHandler{db: db}
}

// healthResponse es el cuerpo que devuelve GET /api/health.
type healthResponse struct {
	Status   string `json:"status"`   // "ok" o "degradado"
	Database string `json:"database"` // "ok", "sin conectar" o "error"
}

// Check maneja GET /api/health.
//   - 200 si la base de datos responde.
//   - 503 (Service Unavailable) si no hay conexión o el ping falla.
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	if h.db == nil {
		writeJSON(w, http.StatusServiceUnavailable, healthResponse{
			Status:   "degradado",
			Database: "sin conectar",
		})
		return
	}

	// Límite de 2 segundos: un healthcheck lento es tan malo como uno fallido.
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	//libera el temporizador interno del contexto cuando la funcion termina, aunque el ping haya
	//respondido en 2ms
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, healthResponse{
			Status:   "degradado",
			Database: "error",
		})
		return
	}

	writeJSON(w, http.StatusOK, healthResponse{Status: "ok", Database: "ok"})
}
