package router

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

// statusRecorder envuelve el ResponseWriter para poder leer, después de que el
// handler terminó, qué código HTTP se respondió (net/http no lo expone).
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// logRequests registra cada petición en una línea de log estructurado:
// método, ruta, código de respuesta y duración.
func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		slog.Info("petición",
			"metodo", r.Method,
			"ruta", r.URL.Path,
			"status", rec.status,
			"duracion_ms", time.Since(start).Milliseconds(),
		)
	})
}

// recoverPanics evita que un panic en un handler tumbe todo el servidor:
// lo registra con su stack trace y responde 500 al cliente.
func recoverPanics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recuperado",
					"error", err,
					"stack", string(debug.Stack()),
				)
				http.Error(w, `{"error":"error interno del servidor"}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
