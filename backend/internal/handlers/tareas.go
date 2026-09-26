package handlers

import (
	"errors"
	"examen-coppel/backend/internal/auth"
	"examen-coppel/backend/internal/httpx"
	"examen-coppel/backend/internal/models"
	"examen-coppel/backend/internal/repository"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// errFechaPasada se devuelve cuando la fecha límite ya pasó.
var errFechaPasada = errors.New("la fecha límite no puede ser anterior a hoy")

// TareaHandler maneja los endpoints del crud de Tareas
type TareaHandler struct {
	repo *repository.TareaRepository
}

// crearTareaRequest es el cuerpo de POST /api/usuarios/{id}/tareas
// el usuario dueño viene en la url, no en el cuerpo
type crearTareaRequest struct {
	tareaRequest
}

// actualizarTareaRequest es el cuerpo de PUT /api/tareas/{id}
// incluye usuario_id cambiarlo es REASIGNAR la tarea
type actualizarTareaRequest struct {
	tareaRequest
	UsuarioID *int   `json:"usuario_id"`
	Estatus   string `json:"estatus"`
}

// estatusRequest es el cuerpo de PATCH /api/tareas/{id}/estatus.
type estatusRequest struct {
	Estatus string `json:"estatus"`
}

// tareaRequest son los campos comunes entre crear y actualizar una tarea
type tareaRequest struct {
	Titulo      string  `json:"titulo"`
	Descripcion *string `json:"descripcion"`
	FechaLimite *string `json:"fecha_limite"` //"AAAA-MM-DD""

	fecha *time.Time
}

// crearTareaGeneralRequest es el cuerpo de POST /api/tareas.
// usuario_id es opcional: si no viene (o es null), la tarea queda sin asignar.
type crearTareaGeneralRequest struct {
	tareaRequest
	UsuarioID *int `json:"usuario_id"`
}

// NewTareaHandler crea el handler con el repositorio de tareas
func NewTareaHandler(repo *repository.TareaRepository) *TareaHandler {
	return &TareaHandler{repo: repo}
}

func (t *tareaRequest) validar() error {
	t.Titulo = strings.TrimSpace(t.Titulo)
	if t.Titulo == "" {
		return errors.New("el título es obligatorio")
	}

	if utf8.RuneCountInString(t.Titulo) > 150 {
		return errors.New("el titulo tiene que ser menor a 150 caracteres")
	}

	if t.Descripcion != nil {
		descripcion := strings.TrimSpace(*t.Descripcion)
		if descripcion == "" {
			t.Descripcion = nil
		} else if utf8.RuneCountInString(descripcion) > 1000 {
			return errors.New("la descripción no puede tener más de 1000 caracteres")
		} else {
			t.Descripcion = &descripcion
		}
	}

	if t.FechaLimite != nil {
		texto := strings.TrimSpace(*t.FechaLimite)
		if texto != "" {
			fecha, err := time.Parse(time.DateOnly, texto)
			if err != nil {
				return errors.New("la fecha límite debe tener el formato AAAA-MM-DD")
			}
			t.fecha = &fecha
		}
	}
	return nil
}

// validar exige usuario_id y estatus (un PUT manda la tarea completa)
// y aplica las reglas comunes de tareaRequest.
func (req *actualizarTareaRequest) validar() error {
	if req.UsuarioID != nil && *req.UsuarioID <= 0 {
		return errors.New("usuario_id inválido")
	}
	// En un PUT el estatus es obligatorio: si viene vacío, esEstatusValido da false.
	if !esEstatusValido(req.Estatus) {
		return errors.New("estatus inválido")
	}
	return req.tareaRequest.validar()
}

// validar exige un estatus válido (no hay valor por defecto).
func (req *estatusRequest) validar() error {
	req.Estatus = strings.TrimSpace(req.Estatus)
	if !esEstatusValido(req.Estatus) {
		return errors.New("estatus inválido: debe ser pendiente, en_progreso o completada")
	}
	return nil
}

// validar aplica las reglas comunes; el estatus por defecto es "pendiente".
func (req *crearTareaRequest) validar() error {
	return req.tareaRequest.validar()
}

func esEstatusValido(estatus string) bool {
	return estatus == models.EstatusCompletada || estatus == models.EstatusEnProgreso || estatus == models.EstatusPendiente
}

// Crear maneja POST /api/usuarios/{id}/tareas: crea una tarea para el usuario {id}.
func (h *TareaHandler) Crear(w http.ResponseWriter, r *http.Request) {
	usuarioID, err := leerID(r) // aquí {id} es el id del USUARIO
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "id inválido")
		return
	}

	var req crearTareaRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
		return
	}
	if err := req.validar(); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.fecha != nil && req.fecha.Before(hoy()) {
		httpx.WriteError(w, http.StatusBadRequest, errFechaPasada.Error())
		return
	}

	tarea := &models.Tarea{
		UsuarioID:   &usuarioID,
		Titulo:      req.Titulo,
		Descripcion: req.Descripcion,
		FechaLimite: req.fecha, // la fecha ya convertida por validar()
		Estatus:     models.EstatusPendiente,
	}

	err = h.repo.Crear(r.Context(), tarea)
	if errors.Is(err, repository.ErrUsuarioNoExiste) {
		httpx.WriteError(w, http.StatusNotFound, "usuario no encontrado")
		return
	}
	if err != nil {
		slog.Error("error al crear tarea", "usuario_id", usuarioID, "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "error interno")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, tarea)
}

// Obtener maneja el GET api/tareas/{id}: devuelve las tareas
// admin puede consultar todas las tareas un usuario solo las suyas
func (h *TareaHandler) Obtener(w http.ResponseWriter, r *http.Request) {
	tarea := h.cargarTareaPropia(w, r)
	if tarea == nil {
		return // el ayudante ya respondió el error
	}
	httpx.WriteJSON(w, http.StatusOK, tarea)
}

// cargarTareaPropia carga la tarea {id} de la URL y verifica que quien pide
// sea admin o su dueño. Si algo falla, ya respondió el error y devuelve nil.
func (h *TareaHandler) cargarTareaPropia(w http.ResponseWriter, r *http.Request) *models.Tarea {

	tareaID, err := leerID(r)

	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "id inválido") //devolvemos el error 400
		return nil
	}

	claims, ok := auth.ClaimsDesdeContexto(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "no autenticado")
		return nil
	}

	tarea, err := h.repo.ObtenerPorID(r.Context(), tareaID)
	if errors.Is(err, repository.ErrNoEncontrado) { // 1. específico
		httpx.WriteError(w, http.StatusNotFound, "tarea no encontrada")
		return nil
	}
	if err != nil { // 2. genérico
		slog.Error("error al obtener tarea", "id", tareaID, "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "error interno")
		return nil
	}
	//tenemos que verificar si el USUARIO puede obtener la tarea solicitada
	//por eso es tarea.UsuarioID
	if !puedeAccederTarea(claims, tarea) {
		//si el usuario no tiene permisos devolvemos 404
		httpx.WriteError(w, http.StatusNotFound, "tarea no encontrada")
		return nil
	}

	return tarea

}

// Actualizar maneja PUT /api/tareas/{id}: modifica una tarea completa (solo admin).
// Cambiar usuario_id reasigna la tarea a otro usuario.
func (h *TareaHandler) Actualizar(w http.ResponseWriter, r *http.Request) {

	// creamos una request para actualizar la tarea
	// no se hace var usuario Models.Tareas porque asi no controlamos lo que manda el usuario, aqui si
	// asi evitamos el mass assigment
	var req actualizarTareaRequest

	tareaID, err := leerID(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "id inválido") //devolvemos el error 400
		return
	}

	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "cuerpo de la peticion invalido")
		return
	}

	err = req.validar()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	tarea, err := h.repo.ObtenerPorID(r.Context(), tareaID)
	if errors.Is(err, repository.ErrNoEncontrado) {
		//si el usuario no existe devolvemos 404
		httpx.WriteError(w, http.StatusNotFound, "tarea no encontrado")
		return
	}
	if err != nil {
		slog.Error("error al obtener la tarea para actualizar", "id", tareaID, "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "error interno")
		return
	}
	// Se permite conservar la fecha que ya tenía, aunque esté vencida.
	if req.fecha != nil && req.fecha.Before(hoy()) && !mismaFecha(req.fecha, tarea.FechaLimite) {
		httpx.WriteError(w, http.StatusBadRequest, errFechaPasada.Error())
		return
	}

	tarea.UsuarioID = req.UsuarioID
	tarea.Titulo = req.Titulo
	tarea.Descripcion = req.Descripcion
	tarea.FechaLimite = req.fecha
	tarea.Estatus = req.Estatus

	err = h.repo.Actualizar(r.Context(), tarea)
	if errors.Is(err, repository.ErrNoEncontrado) {
		httpx.WriteError(w, http.StatusNotFound, "tarea no encontrada") //404
		return
	}

	if errors.Is(err, repository.ErrUsuarioNoExiste) {
		httpx.WriteError(w, http.StatusBadRequest, "el usuario asignado no existe") //400
		return
	}
	if err != nil {
		slog.Error("error al actualizar tarea", "id", tareaID, "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "error") //500
		return
	}
	httpx.WriteJSON(w, http.StatusOK, tarea)
}

// CambiarEstatus maneja PATCH api/tareas/{id}/estatus
// si es admin puede manejar todo, si es usuario solamente las suyas
func (h *TareaHandler) CambiarEstatus(w http.ResponseWriter, r *http.Request) {
	tarea := h.cargarTareaPropia(w, r)
	if tarea == nil {
		return // el ayudante ya respondió el error
	}
	//Creamos nuestra variable estatusRequest para evitar el mass assigment
	var req estatusRequest
	// le pasamos & porque de esa manera le decimos la direccion donde tiene que llenar los datos
	err := httpx.DecodeJSON(w, r, &req)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "cuerpo de la peticion inválido")
		return
	}

	err = req.validar()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	actualizada, err := h.repo.CambiarEstatus(r.Context(), tarea.ID, req.Estatus)
	if errors.Is(err, repository.ErrNoEncontrado) {
		httpx.WriteError(w, http.StatusNotFound, "tarea no encontrada")
		return
	}
	if err != nil {
		slog.Error("error al cambiar el estatus de la tarea", "id", tarea.ID, "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "error interno") //500
		return
	}
	httpx.WriteJSON(w, http.StatusOK, actualizada)
}

// Eliminar controla el DELETE api/tareas/{id}
// solamete el admin es capaz de borrar una tarea
func (h *TareaHandler) Eliminar(w http.ResponseWriter, r *http.Request) {
	tareaID, err := leerID(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "id invalido")
		return
	}

	err = h.repo.Eliminar(r.Context(), tareaID)
	if errors.Is(err, repository.ErrNoEncontrado) {
		httpx.WriteError(w, http.StatusNotFound, "tarea no encontrada")
		return
	}
	if err != nil {
		slog.Error("error al eliminar tarea", "id", tareaID, "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "error interno")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// hoy devuelve la fecha de hoy (según la zona horaria del servidor) a las 00:00 UTC.
// Así se puede comparar con las fechas "AAAA-MM-DD", que time.Parse deja en UTC.
func hoy() time.Time {
	a, m, d := time.Now().Date()
	return time.Date(a, m, d, 0, 0, 0, 0, time.UTC)
}

// mismaFecha dice si dos fechas opcionales son el mismo día.
func mismaFecha(a, b *time.Time) bool {
	return a != nil && b != nil && a.Equal(*b)
}

// Listar maneja GET api/tareas
// solamente el admin puede listar tareas el permiso lo da RequireAdmin en el router
func (h *TareaHandler) Listar(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := repository.FiltroTareas{
		Texto:   strings.TrimSpace(q.Get("q")),
		Estatus: q.Get("estatus"),
	}

	if utf8.RuneCountInString(f.Texto) > 100 {
		httpx.WriteError(w, http.StatusBadRequest, "la búsqueda no puede tener más de 100 caracteres")
		return
	}
	if f.Estatus != "" && !esEstatusValido(f.Estatus) {
		httpx.WriteError(w, http.StatusBadRequest, "estatus inválido")
		return
	}
	if v := q.Get("usuario_id"); v != "" {
		id, err := strconv.Atoi(v)
		if err != nil || id <= 0 {
			httpx.WriteError(w, http.StatusBadRequest, "usuario_id inválido")
			return
		}
		f.UsuarioID = &id
	}

	tareas, err := h.repo.Listar(r.Context(), f)
	if err != nil {
		slog.Error("error al listar tareas", "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "error interno")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, tareas)
}

// CrearGeneral maneja POST /api/tareas: crea una tarea con o sin usuario (solo admin).
func (h *TareaHandler) CrearGeneral(w http.ResponseWriter, r *http.Request) {
	var req crearTareaGeneralRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
		return
	}
	if err := req.validar(); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.fecha != nil && req.fecha.Before(hoy()) {
		httpx.WriteError(w, http.StatusBadRequest, errFechaPasada.Error())
		return
	}

	tarea := &models.Tarea{
		UsuarioID:   req.UsuarioID, // nil = sin asignar
		Titulo:      req.Titulo,
		Descripcion: req.Descripcion,
		FechaLimite: req.fecha,
		Estatus:     models.EstatusPendiente,
	}

	err := h.repo.Crear(r.Context(), tarea)
	if errors.Is(err, repository.ErrUsuarioNoExiste) {
		httpx.WriteError(w, http.StatusBadRequest, "el usuario asignado no existe")
		return
	}
	if err != nil {
		slog.Error("error al crear tarea", "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "error interno")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, tarea)
}

func (r *crearTareaGeneralRequest) validar() error {
	if r.UsuarioID != nil && *r.UsuarioID <= 0 {
		return errors.New("usuario_id inválido")
	}
	return r.tareaRequest.validar()
}

// puedeAccederTarea dice si quien llama puede ver o cambiar esta tarea:
// el admin, cualquiera; un usuario normal, solo las suyas.
// Una tarea sin usuario solo la ve el admin.
func puedeAccederTarea(claims *auth.Claims, t *models.Tarea) bool {
	if t.UsuarioID == nil {
		return claims.Rol == models.RolAdmin
	}
	return puedeAcceder(claims, *t.UsuarioID)
}
