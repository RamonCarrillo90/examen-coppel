package handlers

import (
	"context"
	"errors"
	"examen-coppel/backend/internal/auth"
	"examen-coppel/backend/internal/httpx"
	"examen-coppel/backend/internal/models"
	"examen-coppel/backend/internal/repository"
	"log/slog"
	"net/http"
	"time"
)

// solicitud handler maneja los endpoints del repositorio de solicitudes
type SolicitudHandler struct {
	repo *repository.SolicitudRepository
}

type CrearSolicitudHandler struct {
	ID         int        `json:"id"`
	TareaID    int        `json:"tarea_id"`
	UsuarioID  int        `json:"usuario_id"`
	Estatus    string     `json:"estatus"`
	CreadoEn   time.Time  `json:"creado_en"`
	ResueltoEn *time.Time `json:"resuelto_en"` // nil mientras esté pendiente
}

// NewTareaHandler crea el handler con el repositorio de tareas
func NewSolicitudHandler(repo *repository.SolicitudRepository) *SolicitudHandler {
	return &SolicitudHandler{repo: repo}
}

// esEstatusSolicitudValido indica si estatus es uno de los estatus de solicitud.
func esEstatusSolicitudValido(estatus string) bool {
	return estatus == models.SolicitudPendiente ||
		estatus == models.SolicitudAprobada ||
		estatus == models.SolicitudRechazada
}

// Crear solicitud manejara el endpoint POST /api/tareas/{id}/solicitudes
func (h *SolicitudHandler) Crear(w http.ResponseWriter, r *http.Request) {
	tareaId, err := leerID(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "id invalido")
		return
	}

	claims, ok := auth.ClaimsDesdeContexto(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "no esta autenticado")
		return
	}

	if claims.Rol == models.RolAdmin {
		httpx.WriteError(w, http.StatusBadRequest, "el administrador asigna las tareas directamente")
		return
	}

	usuarioId, err := claims.UsuarioID()

	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "token inválido")
		return
	}

	solicitud, err := h.repo.Crear(r.Context(), tareaId, usuarioId)
	if errors.Is(err, repository.ErrTareaNoDisponible) {
		httpx.WriteError(w, http.StatusConflict, "la tarea ya no está disponible")
		return
	}
	if errors.Is(err, repository.ErrSolicitudDuplicada) {
		httpx.WriteError(w, http.StatusConflict, "ya solicitaste esta tarea")
		return
	}
	if err != nil {
		slog.Error("error al crear solicitud", "tarea_id", tareaId, "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "error interno")
		return
	}
	//si todo salio bien devolvemos 201
	httpx.WriteJSON(w, http.StatusCreated, solicitud)
}

// Mias GET /api/solicitudes/mias
func (h *SolicitudHandler) Mias(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsDesdeContexto(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "no autorizado")
		return
	}

	usuarioID, err := claims.UsuarioID()
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "token invalido")
		return
	}

	f := repository.FiltroSolicitudes{UsuarioID: &usuarioID}
	solicitudes, err := h.repo.Listar(r.Context(), f)
	if err != nil {
		slog.Error("error al listar mis solicitudes", "usuario_id", usuarioID, "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "error interno")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, solicitudes)
}

// Listar maneja GET /api/solicitudes?estatus=...
// Solo para el admin (el permiso lo aplica RequireAdmin en el router).
func (h *SolicitudHandler) Listar(w http.ResponseWriter, r *http.Request) {
	estatus := r.URL.Query().Get("estatus")
	f := repository.FiltroSolicitudes{Estatus: estatus}

	// Vacío = "todas" (válido). Si viene algo, tiene que ser un estatus real.
	if f.Estatus != "" && !esEstatusSolicitudValido(f.Estatus) {
		httpx.WriteError(w, http.StatusBadRequest, "estatus inválido: debe ser pendiente, aprobada o rechazada")
		return
	}

	solicitudes, err := h.repo.Listar(r.Context(), f)
	if err != nil {
		slog.Error("error al listar solicitudes", "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "error interno")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, solicitudes)
}

// Aprobar maneja POST /api/solicitudes/{id}/aprobar (solo admin).
func (h *SolicitudHandler) Aprobar(w http.ResponseWriter, r *http.Request) {
	h.resolver(w, r, h.repo.Aprobar)
}

// Rechazar maneja POST /api/solicitudes/{id}/rechazar (solo admin).
func (h *SolicitudHandler) Rechazar(w http.ResponseWriter, r *http.Request) {
	h.resolver(w, r, h.repo.Rechazar)
}

// resolver tiene la parte común de aprobar y rechazar: leer el id, llamar a la
// acción del repositorio y traducir sus errores a códigos HTTP.
func (h *SolicitudHandler) resolver(w http.ResponseWriter, r *http.Request, accion func(ctx context.Context, id int) error) {
	solicitudID, err := leerID(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "id invalido")
		return
	}

	err = accion(r.Context(), solicitudID)
	if errors.Is(err, repository.ErrNoEncontrado) {
		httpx.WriteError(w, http.StatusNotFound, "solicitud no encontrada")
		return
	}
	if errors.Is(err, repository.ErrSolicitudResuelta) {
		httpx.WriteError(w, http.StatusConflict, "la solicitud ya fue resuelta")
		return
	}
	if errors.Is(err, repository.ErrTareaNoDisponible) {
		httpx.WriteError(w, http.StatusConflict, "la tarea ya fue asignada a otra persona")
		return
	}
	if err != nil {
		slog.Error("error al resolver solicitud", "id", solicitudID, "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "error interno del servidor")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
