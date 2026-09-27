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

// ErrUsuarioNoExiste indica que la tarea hacer referencia a un usuario que no existe
var ErrUsuarioNoExiste = errors.New("el usuario no existe")

type TareaRepository struct {
	db *pgxpool.Pool
}

// Filtro tareas son los filtros opcionales de GET /api/tareas
type FiltroTareas struct {
	Texto      string //buscalo titulo y descripcion
	Estatus    string
	UsuarioID  *int
	SinAsignar bool
}

// NewTareaRepository crea un repositorio que usa el pool de conexiones de db
func NewTareaRepository(db *pgxpool.Pool) *TareaRepository {
	return &TareaRepository{db: db}
}

const columnasTarea = "id, usuario_id, titulo, descripcion, fecha_limite, estatus, creado_en, actualizado_en"

// escanearTarea s scanner es la fila a leer ya se row o rows
// t *models.Tarea es puntero al usuario donde se escriben los datos.
//
//	Si recibiera el struct sin puntero, recibiría una copia, y los datos se perderían al terminar la función.
func escanearTarea(s scanner, t *models.Tarea) error {
	return s.Scan(&t.ID, &t.UsuarioID, &t.Titulo, &t.Descripcion, &t.FechaLimite, &t.Estatus, &t.CreadoEn, &t.ActualizadoEn)
}

// CrearTarea recibe un contexto y un modelo de tipo Tarea
func (r *TareaRepository) Crear(ctx context.Context, t *models.Tarea) error {
	const query = `
		INSERT INTO tareas(usuario_id,titulo,descripcion,fecha_limite,estatus)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id, creado_en, actualizado_en
	`

	err := r.db.QueryRow(ctx, query,
		t.UsuarioID, t.Titulo, t.Descripcion, t.FechaLimite, t.Estatus,
	).Scan(&t.ID, &t.CreadoEn, &t.ActualizadoEn)

	if EsViolacionFK(err) {
		return ErrUsuarioNoExiste
	}
	if err != nil {
		return fmt.Errorf("crear tarea: %w", err)
	}

	return nil
}

func (r *TareaRepository) ObtenerPorID(ctx context.Context, id int) (*models.Tarea, error) {
	query := "SELECT " + columnasTarea + " FROM tareas WHERE id = $1 "

	//modelo de tarea que nos servira para devolver la tarea
	var t models.Tarea
	//llenamos err con escanear tarea para que nos devuelva el modelo
	err := escanearTarea(r.db.QueryRow(ctx, query, id), &t)

	//si no se encuentra ninguna tarea con esa id lanzamos ErrNoEncontrado
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoEncontrado
	}
	// controlamos cualquier otro error y lo envolvemos con %w
	if err != nil {
		return nil, fmt.Errorf("obtener usuario por id: %w", err)
	}
	// si todo sale bien devolvemos la posicion en memoria de t y nil
	return &t, nil
}

// ListarPorUsuario devuelve un arreglo de las tareas listadas por fecha y las que tinen fecha null van hasta el final
func (r *TareaRepository) ListarPorUsuario(ctx context.Context, usuarioID int) ([]models.Tarea, error) {
	query := "SELECT " + columnasTarea + " FROM  tareas WHERE usuario_id = $1 ORDER BY fecha_limite NULLS LAST, id"

	rows, err := r.db.Query(ctx, query, usuarioID)
	if err != nil {
		return nil, fmt.Errorf("listar tareas: %w", err)
	}

	defer rows.Close() //Libera la conexion de vuelta al pool, esta linea se ejecuta hasta que el metodo termine

	tareas := []models.Tarea{} //donde se guardaran las filas que se encuentren

	for rows.Next() {
		var t models.Tarea                              //declaramos una caja vacia para esta fila
		if err := escanearTarea(rows, &t); err != nil { //copia la fila t y revisa errores
			return nil, fmt.Errorf("listar tareas: %w", err)
		}

		tareas = append(tareas, t) //agregamos la nueva tarea a la lista
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("listar tareas: %w", err)
	}
	return tareas, nil
}

// Actualizar actualiza la tarea, puede devolver ErrNoEncontrado si la tarea no existe
// ErrUsuarioNoExiste no existe el usuario
func (r *TareaRepository) Actualizar(ctx context.Context, t *models.Tarea) error {
	const query = `
		UPDATE tareas
		SET usuario_id = $2, titulo = $3, descripcion = $4, fecha_limite = $5, estatus = $6
		WHERE id = $1
		RETURNING actualizado_en`
	//enviamos los datos que queremos actualizar y recibimos actualizado_en

	err := r.db.QueryRow(ctx, query,
		t.ID, t.UsuarioID, t.Titulo, t.Descripcion, t.FechaLimite, t.Estatus,
	).Scan(&t.ActualizadoEn)

	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNoEncontrado
	}
	if EsViolacionFK(err) {
		return ErrUsuarioNoExiste
	}
	if err != nil {
		return fmt.Errorf("actualizar tarea: %w", err)
	}

	return nil
}

func (r *TareaRepository) CambiarEstatus(ctx context.Context, id int, estatus string) (*models.Tarea, error) {
	query := "UPDATE tareas SET estatus = $2 WHERE id = $1 RETURNING " + columnasTarea

	var t models.Tarea //creamos la variable para devolver el modelo de tareas

	err := escanearTarea(r.db.QueryRow(ctx, query, id, estatus), &t)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoEncontrado
	}

	if err != nil {
		return nil, fmt.Errorf("cambiar estatus: %w", err)
	}
	return &t, nil
}

func (r *TareaRepository) Eliminar(ctx context.Context, id int) error {
	const query = "DELETE FROM tareas WHERE id = $1"

	tag, err := r.db.Exec(ctx, query, id)
	if err != nil { // 1. ¿Falló la consulta?
		return fmt.Errorf("eliminar tarea: %w", err)
	}
	if tag.RowsAffected() == 0 { // 2. Funcionó, pero ¿borró algo?
		return ErrNoEncontrado
	}
	return nil
}

func (r *TareaRepository) Listar(ctx context.Context, f FiltroTareas) ([]models.Tarea, error) {
	condiciones := []string{} //creamos la condicion
	args := []any{}

	// Cada filtro agrega su valor a args y usa $N con N = posición en args.
	// El texto del usuario NUNCA se pega en la consulta: siempre viaja como parámetro.
	if f.Texto != "" {
		args = append(args, "%"+f.Texto+"%")
		n := len(args)
		condiciones = append(condiciones, fmt.Sprintf("(titulo ILIKE $%d OR descripcion ILIKE $%d)", n, n))
	}
	if f.Estatus != "" {
		args = append(args, f.Estatus)
		condiciones = append(condiciones, fmt.Sprintf("estatus = $%d", len(args)))
	}
	if f.UsuarioID != nil {
		args = append(args, *f.UsuarioID)
		condiciones = append(condiciones, fmt.Sprintf("usuario_id = $%d", len(args)))
	}
	if f.SinAsignar {
		condiciones = append(condiciones, "usuario_id IS NULL")
	}
	//Aqui hacemos la consulta ya con los filtros
	query := "SELECT " + columnasTarea + " FROM tareas"
	if len(condiciones) > 0 {
		query += " WHERE " + strings.Join(condiciones, " AND ")
	}
	query += " ORDER BY id"

	rows, err := r.db.Query(ctx, query, args...)

	if err != nil {
		return nil, fmt.Errorf("listar tareas: %w", err)
	}
	defer rows.Close() // Para liberar la conexion de vuelta al pool

	tareas := []models.Tarea{} //creamos el arreglo donde meteremos todas nuestras tareas
	for rows.Next() {
		var t models.Tarea                              //creamos el contenedor vacio
		if err := escanearTarea(rows, &t); err != nil { // copia la fila a u y revisa errores
			return nil, fmt.Errorf("listar tareas: %w", err)
		}

		tareas = append(tareas, t) //agregamos la tarea a la lista
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("listar tareas: %w", err)
	}
	return tareas, nil
}
