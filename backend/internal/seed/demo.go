package seed

import (
	"context"
	"errors"
	"examen-coppel/backend/internal/auth"
	"examen-coppel/backend/internal/models"
	"examen-coppel/backend/internal/repository"
	"fmt"
	"log/slog"
	"time"
)

// PasswordDemo es una contraseña para todos los usuarios del ejemplo
// son cuentas de demostracion por eso la contraseña esta en el readme
const PasswordDemo = "demo12345"

// Demo crea usuarios y tareas de ejemplo para que la app no arranque vacia
// es idempotente si el primer usuario ya existe no hace nada
func Demo(ctx context.Context, usuarios *repository.UsuarioRepository, tareas *repository.TareaRepository) error {
	hash, err := auth.HashPassword(PasswordDemo)
	if err != nil {
		return fmt.Errorf("seed demo: %w", err)
	}

	tel := func(s string) *string { return &s }
	nuevos := []*models.Usuario{
		{Nombre: "Ramon", Apellido: "Carrillo", Email: "ramon@demo.com", Telefono: tel("4431608370"), PasswordHash: hash, Rol: models.RolUsuario},
		{Nombre: "Jamzyne", Apellido: "Reyes", Email: "jazmyne@demo.com", Telefono: tel("6871268898"), PasswordHash: hash, Rol: models.RolUsuario},
		{Nombre: "Theo", Apellido: "Carrillo", Email: "theo@demo.com", PasswordHash: hash, Rol: models.RolUsuario},
	}

	for i, u := range nuevos {
		err := usuarios.Crear(ctx, u)
		if i == 0 && errors.Is(err, repository.ErrEmailDuplicado) {
			slog.Info("datos de ejemplo ya existen, se omiten")
			return nil
		}
		if err != nil {
			return fmt.Errorf("seed demo: crear usuario %s: %w", u.Email, err)
		}
	}

	ramon, jazmyne, theo := nuevos[0].ID, nuevos[1].ID, nuevos[2].ID

	// Fechas relativas a hoy, para que nunca aparezcan vencidas en la demo.
	en := func(dias int) *time.Time {
		a, m, d := time.Now().AddDate(0, 0, dias).Date()
		f := time.Date(a, m, d, 0, 0, 0, 0, time.UTC)
		return &f
	}
	texto := func(s string) *string { return &s }

	lista := []*models.Tarea{
		{UsuarioID: &ramon, Titulo: "Preparar reporte mensual de ventas", Descripcion: texto("Ventas por sucursal del mes pasado"), FechaLimite: en(5), Estatus: models.EstatusEnProgreso},
		{UsuarioID: &ramon, Titulo: "Actualizar inventario de bodega", FechaLimite: en(10), Estatus: models.EstatusPendiente},
		{UsuarioID: &ramon, Titulo: "Capacitación de nuevo personal", Estatus: models.EstatusCompletada},
		{UsuarioID: &jazmyne, Titulo: "Revisar pedidos atrasados", Descripcion: texto("Contactar a proveedores con retraso"), FechaLimite: en(3), Estatus: models.EstatusPendiente},
		{UsuarioID: &jazmyne, Titulo: "Cerrar caja del fin de semana", Estatus: models.EstatusCompletada},
		{UsuarioID: &theo, Titulo: "Diseñar promoción de temporada", FechaLimite: en(14), Estatus: models.EstatusEnProgreso},
		// Sin usuario: aparecen en "Tareas disponibles".
		{Titulo: "Organizar archivo de facturas", Descripcion: texto("Digitalizar facturas del trimestre"), FechaLimite: en(7), Estatus: models.EstatusPendiente},
		{Titulo: "Auditoría de precios en piso de venta", FechaLimite: en(12), Estatus: models.EstatusPendiente},
		{Titulo: "Encuesta de satisfacción a clientes", Estatus: models.EstatusPendiente},
	}

	for _, t := range lista {
		if err := tareas.Crear(ctx, t); err != nil {
			return fmt.Errorf("seed demo: crear tarea %q: %w", t.Titulo, err)
		}
	}

	slog.Info("datos de ejemplo creados", "usuarios", len(nuevos), "tareas", len(lista))
	return nil

}
