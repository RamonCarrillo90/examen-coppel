// Package models Define la estructura de los datos de la aplicación
package models

import "time"

// Roles validos de un usuario, deben coincidir con el check de la tabla usuarios
const (
	RolAdmin   = "admin"
	RolUsuario = "usuario"
)

// Usuario representa una fila de la tabla Usuarios
type Usuario struct {
	ID           int     `json:"id"`
	Nombre       string  `json:"nombre"`
	Apellido     string  `json:"apellido"`
	Email        string  `json:"email"`
	Telefono     *string `json:"telefono"` // Utilizamos el puntero ya que la columna admite Null: nil significa "Sin telefono"
	PasswordHash string  `json:"-"`        // encoding/json ignora el campo password_hash ya que este nunca debe salir en las respuestas de la api
	Rol          string  `json:"rol"`
	// time.Time es el equivalente a Timestamptz de postgres
	CreadoEn      time.Time `json:"creado_en"`
	ActualizadoEn time.Time `json:"actualizado_en"`
}
