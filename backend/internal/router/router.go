// Package router define todas las rutas de la API en un solo lugar, para que
// cualquiera que lea el código vea de un vistazo qué endpoints existen.
package router

import (
	"net/http"

	"examen-coppel/backend/internal/handlers"
)

// Handlers agrupa todos los handlers de la API.
type Handlers struct {
	Health  *handlers.HealthHandler
	Auth    *handlers.AuthHandler
	Usuario *handlers.UsuarioHandler
	Tarea   *handlers.TareaHandler
}

// New construye el enrutador HTTP con sus middlewares.
//
// Desde Go 1.22, http.ServeMux entiende el método y los parámetros en la
// ruta ("GET /api/usuarios/{id}"), así que no necesitamos un framework.
func New(h Handlers, jwtSecret []byte) http.Handler {
	mux := http.NewServeMux()

	// Middleware de autenticación, configurado con el secreto (capa 1)
	autenticado := RequireAuth(jwtSecret)

	// --- Sistema -------------------------------------------------------------
	mux.HandleFunc("GET /api/health", h.Health.Check)

	// --- Autenticación  ----------------------------------------------
	mux.HandleFunc("POST /api/auth/login", h.Auth.Login)                     //publico
	mux.Handle("GET /api/auth/me", autenticado(http.HandlerFunc(h.Auth.Me))) // privado

	// --- Usuarios---------------------------------------------------
	//la regla es: sin middlewares handleFunc con middlewares = Handle
	mux.Handle("GET /api/usuarios", autenticado(RequireAdmin(http.HandlerFunc(h.Usuario.Listar))))
	mux.Handle("GET /api/usuarios/{id}", autenticado(http.HandlerFunc(h.Usuario.Obtener)))
	mux.Handle("POST /api/usuarios", autenticado(RequireAdmin(http.HandlerFunc(h.Usuario.Crear))))
	mux.Handle("PUT /api/usuarios/{id}", autenticado(http.HandlerFunc(h.Usuario.Actualizar)))
	mux.Handle("DELETE /api/usuarios/{id}", autenticado(RequireAdmin(http.HandlerFunc(h.Usuario.Eliminar))))

	// --- Tareas ----------------------------------------------------
	mux.Handle("POST /api/usuarios/{id}/tareas", autenticado(RequireAdmin(http.HandlerFunc(h.Tarea.Crear))))
	mux.Handle("GET /api/tareas/{id}", autenticado(http.HandlerFunc(h.Tarea.Obtener)))
	mux.Handle("PUT /api/tareas/{id}", autenticado(RequireAdmin(http.HandlerFunc(h.Tarea.Actualizar))))
	mux.Handle("PATCH /api/tareas/{id}/estatus", autenticado(http.HandlerFunc(h.Tarea.CambiarEstatus)))
	mux.Handle("DELETE /api/tareas/{id}", autenticado(http.HandlerFunc(h.Tarea.Eliminar)))
	mux.Handle("GET /api/tareas", autenticado(RequireAdmin(http.HandlerFunc(h.Tarea.Listar))))
	mux.Handle("POST /api/tareas", autenticado(RequireAdmin(http.HandlerFunc(h.Tarea.CrearGeneral))))
	mux.Handle("GET /api/tareas/disponibles", autenticado(http.HandlerFunc(h.Tarea.Disponibles)))
	// Los middlewares se aplican de afuera hacia adentro:
	// recoverPanics -> logRequests -> mux
	return recoverPanics(logRequests(mux))
}
