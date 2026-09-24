// Package router define todas las rutas de la API en un solo lugar, para que
// cualquiera que lea el código vea de un vistazo qué endpoints existen.
package router

import (
	"net/http"

	"examen-coppel/backend/internal/handlers"
)

// New construye el enrutador HTTP con sus middlewares.
//
// Desde Go 1.22, http.ServeMux entiende el método y los parámetros en la
// ruta ("GET /api/usuarios/{id}"), así que no necesitamos un framework.
func New(healthHandler *handlers.HealthHandler, authHandler *handlers.AuthHandler, usuarioHandler *handlers.UsuarioHandler, jwtSecret []byte) http.Handler {

	mux := http.NewServeMux()

	// Middleware de autenticación, configurado con el secreto (capa 1)
	autenticado := RequireAuth(jwtSecret)

	// --- Sistema -------------------------------------------------------------
	mux.HandleFunc("GET /api/health", healthHandler.Check)

	// --- Autenticación (jueves) ----------------------------------------------
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)                     //publico
	mux.Handle("GET /api/auth/me", autenticado(http.HandlerFunc(authHandler.Me))) // privado

	// --- Usuarios (jueves) ---------------------------------------------------
	//la regla es: sin middlewares handleFunc con middlewares = Handle
	mux.Handle("GET /api/usuarios", autenticado(RequireAdmin(http.HandlerFunc(usuarioHandler.Listar))))
	mux.Handle("GET /api/usuarios/{id}", autenticado(http.HandlerFunc(usuarioHandler.Obtener)))
	mux.Handle("POST /api/usuarios", autenticado(RequireAdmin(http.HandlerFunc(usuarioHandler.Crear))))
	mux.Handle("PUT /api/usuarios/{id}", autenticado(http.HandlerFunc(usuarioHandler.Actualizar)))
	mux.Handle("DELETE /api/usuarios/{id}", autenticado(RequireAdmin(http.HandlerFunc(usuarioHandler.Eliminar))))

	// --- Tareas (viernes) ----------------------------------------------------
	// mux.HandleFunc("POST /api/usuarios/{id}/tareas", ...)
	// mux.HandleFunc("GET /api/tareas/{id}", ...)
	// mux.HandleFunc("PUT /api/tareas/{id}", ...)
	// mux.HandleFunc("DELETE /api/tareas/{id}", ...)

	// Los middlewares se aplican de afuera hacia adentro:
	// recoverPanics -> logRequests -> mux
	return recoverPanics(logRequests(mux))
}
