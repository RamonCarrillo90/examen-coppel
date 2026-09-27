package repository

import (
	"context"
	"errors"
	"examen-coppel/backend/internal/models"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	// ErrSolicitudDuplicada: el usuario ya tiene una solicitud pendiente para esa tarea.
	ErrSolicitudDuplicada = errors.New("ya existe una solicitud pendiente")
	// ErrTareaNoDisponible: la tarea no existe o ya tiene usuario asignado.
	ErrTareaNoDisponible = errors.New("la tarea no está disponible")
	// ErrSolicitudResuelta: la solicitud ya fue aprobada o rechazada.
	ErrSolicitudResuelta = errors.New("la solicitud ya fue resuelta")
)

// SolicitudRepository guarda y consulta las solicitudes de tareas.
type SolicitudRepository struct {
	db *pgxpool.Pool
}

// NewSolicitudRepository crea el repositorio con el pool de conexiones.
func NewSolicitudRepository(db *pgxpool.Pool) *SolicitudRepository {
	return &SolicitudRepository{db: db}
}

// FiltroSolicitudes son los filtros opcionales al listar.
type FiltroSolicitudes struct {
	Estatus   string
	UsuarioID *int // solo las de este usuario ("mis solicitudes")
}

// escanearSolicitud s scanner es la fila a leer ya se row o rows
// sl *models.Solicitud es puntero al usuario donde se escriben los datos.
//
//	Si recibiera el struct sin puntero, recibiría una copia, y los datos se perderían al terminar la función.
func escanearSolicitud(s scanner, sl *models.SolicitudDetalle) error {
	return s.Scan(&sl.ID, &sl.TareaID, &sl.UsuarioID, &sl.Estatus, &sl.CreadoEn, &sl.ResueltoEn, &sl.TareaTitulo, &sl.UsuarioNombre)
}

func (r *SolicitudRepository) Crear(ctx context.Context, tareaID int, usuarioID int) (*models.Solicitud, error) {
	const query = `
		INSERT INTO solicitudes(tarea_id, usuario_id)
		SELECT id,$2 FROM tareas WHERE id = $1 and usuario_id IS NULL
		RETURNING id, tarea_id, usuario_id, estatus, creado_en, resuelto_en
	`
	var s models.Solicitud //Creamos una variable para asi recibir los datos que vamos a escanear
	err := r.db.QueryRow(ctx, query, tareaID, usuarioID).Scan(&s.ID, &s.TareaID, &s.UsuarioID, &s.Estatus, &s.CreadoEn, &s.ResueltoEn)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTareaNoDisponible
	}
	if EsUniqueViolation(err) { //verificamos que no pueda haber solicitudes duplicadas
		return nil, ErrSolicitudDuplicada
	}
	if err != nil {
		return nil, fmt.Errorf("crear solicitud: %w", err)
	}

	return &s, nil
}

func (r *SolicitudRepository) Listar(ctx context.Context, f FiltroSolicitudes) ([]models.SolicitudDetalle, error) {
	//Armamos el query
	query := `
		SELECT s.id, s.tarea_id, s.usuario_id, s.estatus, s.creado_en, s.resuelto_en,
        t.titulo, u.nombre || ' ' || u.apellido
		FROM solicitudes s
		JOIN tareas t   ON t.id = s.tarea_id
		JOIN usuarios u ON u.id = s.usuario_id`

	condiciones := []string{}
	args := []any{}

	if f.Estatus != "" {
		args = append(args, f.Estatus)
		condiciones = append(condiciones, fmt.Sprintf("s.estatus = $%d", len(args)))
	}

	if f.UsuarioID != nil {
		args = append(args, *f.UsuarioID)
		condiciones = append(condiciones, fmt.Sprintf("s.usuario_id = $%d", len(args)))
	}

	if len(condiciones) > 0 {
		query += " WHERE " + strings.Join(condiciones, " AND ")
	}
	query += " ORDER BY s.creado_en DESC, s.id DESC"

	// Ejecutamos la consulta
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listar solicitudes: %w", err)
	}

	defer rows.Close() //Devolvemos la conexion al pool

	solicitudes := []models.SolicitudDetalle{}
	for rows.Next() {
		var s models.SolicitudDetalle
		if err := escanearSolicitud(rows, &s); err != nil {
			return nil, fmt.Errorf("listar solicitudes: %w", err)
		}
		solicitudes = append(solicitudes, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("listar solicitudes: %w", err)
	}

	return solicitudes, nil
}

func (r *SolicitudRepository) Aprobar(ctx context.Context, id int) error {
	//empezamos la transaccion en pgx
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("aprobar solicitud: %w", err)
	}
	defer tx.Rollback(ctx)
	//primer query leer y bloquear la solicitud
	var (
		tareaID, usuarioID int
	)
	var estatus string

	query := `SELECT tarea_id, usuario_id, estatus FROM solicitudes WHERE id = $1 FOR UPDATE`
	// id es el id de la solicitud
	err = tx.QueryRow(ctx, query, id).Scan(&tareaID, &usuarioID, &estatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNoEncontrado
	}
	if err != nil {
		return fmt.Errorf("aprobar solicitud: %w", err)
	}
	if estatus != models.SolicitudPendiente {
		return ErrSolicitudResuelta
	}

	//segundo query, asignar la tarea solo si sigue libre
	query = `UPDATE tareas SET usuario_id = $1, actualizado_en = now() WHERE id = $2 AND usuario_id IS NULL`
	tag, err := tx.Exec(ctx, query, usuarioID, tareaID)
	if err != nil {
		return fmt.Errorf("aprobar solicitud: %w", err)
	}

	if tag.RowsAffected() <= 0 {
		return ErrTareaNoDisponible
	}

	// tercer query marcar la solicitud como aprobada
	query = `UPDATE solicitudes SET estatus = 'aprobada', resuelto_en = now() WHERE id = $1`
	_, err = tx.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("aprobar solicitud: %w", err)
	}

	//cuarto query rechazar las demas pendientes de esa tarea
	query = `UPDATE solicitudes SET estatus = 'rechazada', resuelto_en = now() 
		WHERE tarea_id = $1 AND estatus = 'pendiente' AND id <> $2`

	_, err = tx.Exec(ctx, query, tareaID, id)
	if err != nil {
		return fmt.Errorf("aprobar solicitud: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("aprobar solicitud: %w", err)
	}
	return nil
}

func (r *SolicitudRepository) Rechazar(ctx context.Context, id int) error {
	query := `UPDATE solicitudes SET estatus = 'rechazada', resuelto_en = now()
		WHERE id = $1 AND estatus = 'pendiente'`

	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("rechazar solicitud: %w", err)
	}

	if tag.RowsAffected() > 0 {
		return nil
	}
	//si el resultado del tag fue 0 no sabemos si existe o si ya estaba resuelta
	//asi que lo consultamos
	query = `SELECT EXISTS (SELECT 1 FROM solicitudes WHERE id = $1)`

	var existe bool

	err = r.db.QueryRow(ctx, query, id).Scan(&existe)
	if err != nil {
		return fmt.Errorf("rechazar solicitud: %w", err)
	}
	if !existe {
		return ErrNoEncontrado
	}

	return ErrSolicitudResuelta

}
