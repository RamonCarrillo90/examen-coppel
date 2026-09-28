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
	_ "time/tzdata" // incluye la base de zonas horarias dentro del binario

	"examen-coppel/backend/internal/config"
	"examen-coppel/backend/internal/database"
	"examen-coppel/backend/internal/handlers"
	"examen-coppel/backend/internal/repository"
	"examen-coppel/backend/internal/router"
	"examen-coppel/backend/internal/seed"
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

	// 1. Conexión a la base de datos
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	slog.Info("conectado a PostgreSQL")

	// 2. Repositorios (acceso a datos)
	usuarioRepo := repository.NewUsuarioRepository(pool)
	tareaRepo := repository.NewTareaRepository(pool)
	solicitudRepo := repository.NewSolicitudRepository(pool)

	// 3. Admin inicial
	if err := seed.CrearAdminSiNoExiste(ctx, usuarioRepo, cfg.AdminEmail, cfg.AdminPassword); err != nil {
		return err
	}
	if os.Getenv("SEED_DEMO") == "true" {
		if err := seed.Demo(ctx, usuarioRepo, tareaRepo); err != nil {
			slog.Error("no se pudieron crear los datos de ejemplo", "error", err)
			os.Exit(1)
		}
	}

	// 4. Handlers (HTTP)
	jwtSecret := []byte(cfg.JWTSecret)

	healthHandler := handlers.NewHealthHandler(pool)
	authHandler := handlers.NewAuthHandler(usuarioRepo, jwtSecret)
	usuarioHandler := handlers.NewUsuarioHandler(usuarioRepo, tareaRepo)
	tareaHandler := handlers.NewTareaHandler(tareaRepo)
	solicitudHandler := handlers.NewSolicitudHandler(solicitudRepo)

	// 5. Servidor con el router
	srv := &http.Server{
		Addr: ":" + cfg.Port,
		Handler: router.New(router.Handlers{
			Health:    healthHandler,
			Auth:      authHandler,
			Usuario:   usuarioHandler,
			Tarea:     tareaHandler,
			Solicitud: solicitudHandler,
		}, jwtSecret),
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
