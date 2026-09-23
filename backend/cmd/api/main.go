// Punto de entrada de la API REST de usuarios y tareas.
//
// Responsabilidades de main (y nada más):
//  1. Cargar la configuración.
//  2. Crear las dependencias (conexión a BD, handlers, router).
//  3. Arrancar el servidor HTTP y apagarlo limpiamente al recibir una señal.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"examen-coppel/backend/internal/config"
	"examen-coppel/backend/internal/database"
	"examen-coppel/backend/internal/handlers"
	"examen-coppel/backend/internal/router"
)

func main() {
	// Logs estructurados en JSON: fáciles de leer para humanos y de filtrar
	// para herramientas (docker logs, Grafana Loki, etc.).
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	if err := run(); err != nil {
		slog.Error("la aplicación terminó con error", "error", err)
		os.Exit(1)
	}
}

// run contiene la lógica real de arranque. Separarla de main permite devolver
// errores de forma normal en lugar de llamar os.Exit en cada punto.
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// ctx se cancela cuando llega Ctrl+C (SIGINT) o `docker stop` (SIGTERM).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// -------------------------------------------------------------------------
	// PASO FINAL (tu parte): descomenta este bloque cuando hayas escrito
	// database.Connect, y cambia NewHealthHandler(nil) por NewHealthHandler(pool).
	// No olvides agregar "examen-coppel/backend/internal/database" a los imports.
	// -------------------------------------------------------------------------
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	slog.Info("conectado a PostgreSQL")

	healthHandler := handlers.NewHealthHandler(pool)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router.New(healthHandler),
		// Timeouts: sin ellos, un cliente lento puede mantener conexiones
		// abiertas indefinidamente y agotar los recursos del servidor al hacer un ataque slowloris
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// El servidor corre en una goroutine para que main pueda quedarse
	// esperando la señal de apagado.
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("servidor escuchando", "puerto", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	// Esperamos lo que ocurra primero: un error del servidor o una señal.
	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		slog.Info("señal de apagado recibida, cerrando servidor...")
	}

	// Graceful shutdown: deja de aceptar conexiones nuevas y da hasta 10 s
	// a las peticiones en curso para terminar.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	slog.Info("servidor detenido correctamente")
	return nil
}
