// Package repository encapsula el acceso a postgreSQL es la unica capa de la aplicacion que ejecuta Sql
package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"examen-coppel/backend/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Errores que devuelve el repositorio, los handlers los comparan con Error.Is
// Y asi determinan el codigo HTTP del error
var (
	// ErrNoEncontrado Indica que no existe un registro con el criterio buscado
	ErrNoEncontrado = errors.New("registro no encontrado")
	// ErrEmailDuplicado Indica que el email que se ingresa ya existe
	ErrEmailDuplicado = errors.New("el email ya esta registrado")
)

// UsuarioRepository ejecuta las consultas SQL de la tabla de usuarios
// Es seguro usarlo desde varias goroutines, porque el pool maneja la concurrencia.
type UsuarioRepository struct {
	db *pgxpool.Pool
}

type FiltroUsuarios struct {
	Texto string
	Rol   string
}

// NewUsuarioRepository crea un repositorio que usa el pool de conexiones de db
func NewUsuarioRepository(db *pgxpool.Pool) *UsuarioRepository {
	return &UsuarioRepository{db: db}
}

// columnasUsuario es la lista de columnas que leemos en cada Select/Returning
// en el mismo orden en que escanearUsuario las copia al return
const columnasUsuario = `id, nombre, apellido, email, telefono, password_hash, rol, creado_en, actualizado_en`

// scanner es cualquier cosa con Scan: pgx.row (una fila) o pgx.rows (varias)
type scanner interface {
	Scan(dest ...any) error
}

// escanearUsuario s scanner es la fila a leer ya se row o rows
// u *models.Usuario es puntero al usuario donde se escriben los datos.
//
//	Si recibiera el struct sin puntero, recibiría una copia, y los datos se perderían al terminar la función.
func escanearUsuario(s scanner, u *models.Usuario) error {
	return s.Scan(&u.ID, &u.Nombre, &u.Apellido, &u.Email, &u.Telefono, &u.PasswordHash, &u.Rol, &u.CreadoEn, &u.ActualizadoEn)
}

// Crear es una fucion que recibe un usuario y lo inserta en la tabla y guarda en el struct lo que el postgres genero: el id y las fechas
// en go r hace el papel de this
func (r *UsuarioRepository) Crear(ctx context.Context, u *models.Usuario) error {
	//guardamos el query en una constante
	const query = `
		INSERT INTO usuarios(nombre, apellido, email, telefono, password_hash ,rol)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, creado_en, actualizado_en
	`

	err := r.db.QueryRow(ctx, query,
		u.Nombre, u.Apellido, u.Email, u.Telefono, u.PasswordHash, u.Rol,
	).Scan(&u.ID, &u.CreadoEn, &u.ActualizadoEn)

	if esEmailDuplicado(err) {
		return ErrEmailDuplicado
	}
	if err != nil {
		return fmt.Errorf("crear usuario: %w", err)
	}

	return nil
}

// obtenerPorID tiene un puntero apuntando a un usuario porque eso es lo que va a devolver
func (r *UsuarioRepository) ObtenerPorID(ctx context.Context, id int) (*models.Usuario, error) {
	//No declaramos un const porque al utilizar columnasUsuario deja de ser constante y empieza a ser dinamico
	query := "SELECT " + columnasUsuario + " FROM  usuarios WHERE id = $1"
	// modelo de usuario que nos servira para devolver el usuario que coincida con nuestra id
	var u models.Usuario

	// err se llena con escanearUsuario para asi filtrar por id
	err := escanearUsuario(r.db.QueryRow(ctx, query, id), &u)
	//si no se encuentra ningun usuario con esa id lanzamos ErrNoEncontrado
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoEncontrado
	}
	// controlamos cualquier otro error y lo envolvemos con %w
	if err != nil {
		return nil, fmt.Errorf("obtener usuario por id: %w", err)
	}
	// si todo sale bien devolvemos la posicion en memoria de u y nil
	return &u, nil
}

// ObtenerPorEmail nos devolvera el usuario que coincida con el email que solicitemos
func (r *UsuarioRepository) ObtenerPorEmail(ctx context.Context, email string) (*models.Usuario, error) {
	//No declaramos un const porque al utilizar columnasUsuario deja de ser constante y empieza a ser dinamico
	query := "SELECT " + columnasUsuario + " FROM  usuarios WHERE email = $1"
	// modelo de usuario que nos servira para devolver el usuario que coincida con nuestro email
	var u models.Usuario

	// err se llena con escanearUsuario para asi filtrar por email
	err := escanearUsuario(r.db.QueryRow(ctx, query, email), &u)
	//si no se encuentra ningun usuario con este email lanzamos ErrNoEncontrado
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoEncontrado
	}
	// controlamos cualquier otro error y lo envolvemos con %w
	if err != nil {
		return nil, fmt.Errorf("obtener usuario por email: %w", err)
	}
	// si todo sale bien devolvemos la posicion en memoria de u y nil
	return &u, nil
}

func (r *UsuarioRepository) Listar(ctx context.Context, f FiltroUsuarios) ([]models.Usuario, error) {
	condiciones := []string{}
	args := []any{}

	if f.Texto != "" {
		args = append(args, "%"+f.Texto+"%")
		n := len(args)
		condiciones = append(condiciones, fmt.Sprintf(
			"(nombre ILIKE $%d OR apellido ILIKE $%d OR email ILIKE $%d)", n, n, n))
	}
	if f.Rol != "" {
		args = append(args, f.Rol)
		condiciones = append(condiciones, fmt.Sprintf("rol = $%d", len(args)))
	}
	// Cada filtro agrega su valor a args y usa $N con N = posición en args.
	// El texto del usuario NUNCA se pega en la consulta: siempre viaja como parámetro.

	query := "SELECT " + columnasUsuario + " FROM usuarios"
	if len(condiciones) > 0 {
		query += " WHERE " + strings.Join(condiciones, " AND ")
	}
	query += " ORDER BY id"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error al listar los usuarios: %w", err)
	}
	defer rows.Close() //Libera la conexion de vuelta al pool, esta linea se ejecuta hasta que el metodo termine

	usuarios := []models.Usuario{} // vacio, no nil, se serializa como [] y no como null
	for rows.Next() {
		var u models.Usuario                              // caja vacia para esta fila
		if err := escanearUsuario(rows, &u); err != nil { // copia la fila a u y revisa errores
			return nil, fmt.Errorf("listar usuarios: %w", err)
		}

		usuarios = append(usuarios, u) //agregamos el nuevo usuario a la lista
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("listar usuarios: %w", err)
	}
	return usuarios, nil
}

// Actualizar actualiza la informacion de un usuario
func (r *UsuarioRepository) Actualizar(ctx context.Context, u *models.Usuario) error {
	const query = `
		UPDATE usuarios
		SET nombre = $2, apellido = $3, email = $4, telefono = $5, rol = $6
		WHERE id = $1
		RETURNING actualizado_en`
	//Enviamos los datos que queremos actualizar y recibimos actualizadoEn desde la db
	err := r.db.QueryRow(ctx, query,
		u.ID, u.Nombre, u.Apellido, u.Email, u.Telefono, u.Rol,
	).Scan(&u.ActualizadoEn)
	// si no existe la id lanzamos el error
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNoEncontrado
	}
	// esEmailDuplicado verifica el campo  unique del email
	if esEmailDuplicado(err) {
		return ErrEmailDuplicado
	}
	//Preveemos cualquier otro error
	if err != nil {
		return fmt.Errorf("actualizar usuario: %w", err)
	}

	return nil
}

// CambiarPassword Actualiza la contraseña del usuario
func (r *UsuarioRepository) CambiarPassword(ctx context.Context, id int, hash string) error {
	const query = `
	UPDATE usuarios SET password_hash = $2 WHERE id = $1
	`
	// Utilizamos db.Exec porque no hay ningun returning entonces no es necesario
	tag, err := r.db.Exec(ctx, query,
		id, hash)
	// Si no hay ninguna fila afectada entonces el usuario no existia
	if tag.RowsAffected() == 0 {
		return ErrNoEncontrado
	}
	// Envolvemos el error con mas contexto
	if err != nil {
		return fmt.Errorf("cambiar password: %w", err)
	}

	return nil

}

// Eliminar elimina usuarios de la base de datos
func (r *UsuarioRepository) Eliminar(ctx context.Context, id int) error {
	const query = `DELETE FROM usuarios WHERE id = $1`
	// Utilizamos db.Exec porque no hay ningun returning entonces no es necesario
	tag, err := r.db.Exec(ctx, query, id)
	// Envolvemos el error con mas contexto
	if err != nil {
		return fmt.Errorf("eliminar usuario: %w", err)
	}
	// Si no hay ninguna fila afectada entonces el usuario no existia
	if tag.RowsAffected() == 0 {
		return ErrNoEncontrado
	}

	return nil
}

// esEmailDuplicado indica si err es una violacion de la condicion UNIQUE
// ponemos el codigo 23505 de postgreSQL es decir un email ya registrado
func esEmailDuplicado(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
