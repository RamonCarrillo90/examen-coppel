package models

import "time"

const (
	EstatusPendiente  = "pendiente"
	EstatusEnProgreso = "en_progreso"
	EstatusCompletada = "completada"
)

type Tarea struct {
	ID          int        `json:"id"`
	UsuarioID   *int       `json:"usuario_id"`
	Titulo      string     `json:"titulo"`
	Descripcion *string    `json:"descripcion"`  // usamos un puntero ya que la columna admite NULL: esto significa que puede no llevar descripcion
	FechaLimite *time.Time `json:"fecha_limite"` // igual aqui puede no tener fecha limite
	Estatus     string     `json:"estatus"`
	// time.Time es el equivalente a Timestamptz de postgres
	CreadoEn      time.Time `json:"creado_en"`
	ActualizadoEn time.Time `json:"actualizado_en"`
}
