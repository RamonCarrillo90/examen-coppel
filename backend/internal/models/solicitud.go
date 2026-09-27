package models

import "time"

// Estatus posibles de una solicitud.
const (
	SolicitudPendiente = "pendiente"
	SolicitudAprobada  = "aprobada"
	SolicitudRechazada = "rechazada"
)

// Solicitud es la petición de un usuario para quedarse con una tarea sin asignar.
type Solicitud struct {
	ID         int        `json:"id"`
	TareaID    int        `json:"tarea_id"`
	UsuarioID  int        `json:"usuario_id"`
	Estatus    string     `json:"estatus"`
	CreadoEn   time.Time  `json:"creado_en"`
	ResueltoEn *time.Time `json:"resuelto_en"` // nil mientras esté pendiente
}

// SolicitudDetalle agrega los datos que la pantalla necesita mostrar
// (título de la tarea y nombre del solicitante), para no pedirlos aparte.
type SolicitudDetalle struct {
	Solicitud
	TareaTitulo   string `json:"tarea_titulo"`
	UsuarioNombre string `json:"usuario_nombre"`
}
