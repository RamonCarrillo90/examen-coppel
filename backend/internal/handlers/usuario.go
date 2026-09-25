package handlers

import (
	"errors"
	"examen-coppel/backend/internal/auth"
	"examen-coppel/backend/internal/httpx"
	"examen-coppel/backend/internal/models"
	"examen-coppel/backend/internal/repository"
	"log/slog"
	"net/http"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// telefonoValido: exactamente 10 dígitos, sin espacios ni guiones.
var telefonoValido = regexp.MustCompile(`^[0-9]{10}$`)

// UsuarioHandler maneja los endpoints del crud de usuarios
type UsuarioHandler struct {
	repo      *repository.UsuarioRepository
	tareaRepo *repository.TareaRepository
}

// crearUsuarioRequest es el cuerpo de POST /api/usuarios.
type crearUsuarioRequest struct {
	datosUsuario
	Password string `json:"password"`
	Rol      string `json:"rol"`
}

// actualizarUsuarioRequest es el cuerpo de PUT /api/usuarios/{id}.
// Password y Rol son opcionales: si vienen vacíos, se conserva el valor actual.
type actualizarUsuarioRequest struct {
	datosUsuario
	Password string `json:"password"`
	Rol      string `json:"rol"`
}

// datosUsuario son los campos comunes de crear y actualizar un usuario.
type datosUsuario struct {
	Nombre   string  `json:"nombre"`
	Apellido string  `json:"apellido"`
	Email    string  `json:"email"`
	Telefono *string `json:"telefono"`
}

// usuarioConTareas es la respuesta de GET /api/usuarios/{id}:
// los datos del usuario y, además, la lista de sus tareas.
type usuarioConTareas struct {
	*models.Usuario
	Tareas []models.Tarea `json:"tareas"`
}

// NewUsuarioHandler crea el handler con el repositorio de usuarios
func NewUsuarioHandler(repo *repository.UsuarioRepository, tareaRepo *repository.TareaRepository) *UsuarioHandler {
	return &UsuarioHandler{repo: repo, tareaRepo: tareaRepo}
}

// puedeAcceder indica si quien hace la petición puede ver o editar al usuario id:
// los admins pueden con cualquiera, los usuarios normales solo consigo mismos.
func puedeAcceder(claims *auth.Claims, id int) bool {
	//los admins pueden con cualquiera
	if claims.Rol == models.RolAdmin {
		return true
	}

	//Usuario normal ¿El id de su token es el mismo que el de la url?
	idToken, err := claims.UsuarioID()

	if err != nil {
		return false //si el token trae un id raro por seguridad no
	}

	return idToken == id
}

// leerID obtiene el id de la url y lo convierte a un entero postivo
func leerID(r *http.Request) (int, error) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		return 0, errors.New("id invalido")
	}
	return id, nil
}

// validar limpia (espacios, minúsculas) y valida los datos personales.
func (d *datosUsuario) validar() error {
	d.Nombre = strings.TrimSpace(d.Nombre)
	d.Apellido = strings.TrimSpace(d.Apellido)
	if d.Nombre == "" || d.Apellido == "" {
		return errors.New("el nombre y el apellido son obligatorios")
	}
	if utf8.RuneCountInString(d.Nombre) > 100 || utf8.RuneCountInString(d.Apellido) > 100 {
		return errors.New("el nombre y el apellido no pueden tener más de 100 caracteres")
	}

	d.Email = strings.ToLower(strings.TrimSpace(d.Email))
	if utf8.RuneCountInString(d.Email) > 150 {
		return errors.New("el email no puede tener más de 150 caracteres")
	}
	if !esEmailValido(d.Email) {
		return errors.New("el email no tiene un formato válido")
	}

	// El teléfono es opcional; si viene, debe tener 10 dígitos (formato de México).
	if d.Telefono != nil {
		t := strings.TrimSpace(*d.Telefono)
		if t == "" {
			d.Telefono = nil // vacío = sin teléfono
		} else if !telefonoValido.MatchString(t) {
			return errors.New("el teléfono debe tener 10 dígitos")
		} else {
			d.Telefono = &t
		}
	}
	return nil
}

// validarPassword aplica las reglas de longitud de contraseña.
// El máximo se mide en bytes porque bcrypt solo acepta hasta 72 bytes.
func validarPassword(password string) error {
	if utf8.RuneCountInString(password) < 8 {
		return errors.New("la contraseña debe tener al menos 8 caracteres")
	}
	if len(password) > 72 {
		return errors.New("la contraseña es demasiado larga")
	}
	return nil
}

// esRolValido indica si rol es uno de los roles permitidos.
func esRolValido(rol string) bool {
	return rol == models.RolAdmin || rol == models.RolUsuario
}

// validar exige contraseña y asigna el rol "usuario" si no viene.
func (req *crearUsuarioRequest) validar() error {
	if err := req.datosUsuario.validar(); err != nil {
		return err
	}
	if err := validarPassword(req.Password); err != nil {
		return err
	}
	if req.Rol == "" {
		req.Rol = models.RolUsuario
	}
	if !esRolValido(req.Rol) {
		return errors.New("rol inválido: debe ser admin o usuario")
	}
	return nil
}

// validar solo revisa la contraseña y el rol si vienen.
func (req *actualizarUsuarioRequest) validar() error {
	if err := req.datosUsuario.validar(); err != nil {
		return err
	}
	if req.Password != "" {
		if err := validarPassword(req.Password); err != nil {
			return err
		}
	}
	if req.Rol != "" && !esRolValido(req.Rol) {
		return errors.New("rol inválido: debe ser admin o usuario")
	}
	return nil
}

func esEmailValido(email string) bool {
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email
}

func esSoloNumeros(s string) bool {
	for _, caracter := range s {
		if caracter < '0' || caracter > '9' {
			return false // Encontró una letra o símbolo
		}
	}
	return true
}

// Listar maneja GET /api/usuarios: devuelve todos los usuarios.
// Solo para admins; el permiso lo aplica RequireAdmin en el router.
func (h *UsuarioHandler) Listar(w http.ResponseWriter, r *http.Request) {
	usuarios, err := h.repo.Listar(r.Context()) //r.Context(): si el cliente cancela la petición, la consulta también se cancela.
	if err != nil {
		slog.Error("error al listar usuarios", "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "error interno")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, usuarios)
}

// Obtener maneja el GET /api/usuarios/{id}: devuelve un usuario por id
// Admin puede ver cualquiera y el usuario normal solamente el suyo
func (h *UsuarioHandler) Obtener(w http.ResponseWriter, r *http.Request) {

	id, err := leerID(r)
	if err != nil {
		// si leerID falla devolvemos 400
		httpx.WriteError(w, http.StatusBadRequest, "el id no es válido")
		return //return para que si un usuario no autenticado intenta algo que no puede el return lo saque
	}

	claims, ok := auth.ClaimsDesdeContexto(r.Context())

	if !ok {
		//si el usuario no esta autorizado usamos el 401
		httpx.WriteError(w, http.StatusUnauthorized, "no autenticado")
		return
	}

	if !puedeAcceder(claims, id) {
		//si el usuario no tiene permisos devolvemos 403
		httpx.WriteError(w, http.StatusForbidden, "no tienes permisos para esta acción")
		return
	}

	usuario, err := h.repo.ObtenerPorID(r.Context(), id)
	if errors.Is(err, repository.ErrNoEncontrado) {
		//si el usuario no existe devolvemos 404
		httpx.WriteError(w, http.StatusNotFound, "usuario no encontrado")
		return
	}

	if err != nil {
		//registramos el error real antes del 500 para si algo llega a fallar saber que paso
		slog.Error("error al obtener usuario", "id", id, "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "error interno del servidor")
		return
	}

	tareas, err := h.tareaRepo.ListarPorUsuario(r.Context(), id)

	if err != nil {
		slog.Error("error al listar tareas del usuario", "id", id, "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "error interno")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, usuarioConTareas{Usuario: usuario, Tareas: tareas})

}

// Crer maneja el POST /api/usuarios que se encarga de crear un usuario
// el admin es el unico que puede crear usuarios y si un usuario lo intenta sale 403
func (h *UsuarioHandler) Crear(w http.ResponseWriter, r *http.Request) {
	//creamos una request para crear usuario
	//no se hace var usuario Models.Usuario porque asi no controlamos lo que manda el usuario, aqui si
	var req crearUsuarioRequest

	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "cuerpo de la peticion invalido")
		return
	}

	err := req.validar()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	//hasheamos la contraseña
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		slog.Error("error al hashear contraseña", "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "error interno del servidor")
		return
	}
	//armamos el modelo campo por campo
	usuario := &models.Usuario{
		Nombre:       req.Nombre,
		Apellido:     req.Apellido,
		Email:        req.Email,
		Telefono:     req.Telefono,
		PasswordHash: hash, // no lo tiene el cliente, lo calculamos con HashPasswordS
		Rol:          req.Rol,
	}

	err = h.repo.Crear(r.Context(), usuario)
	if errors.Is(err, repository.ErrEmailDuplicado) {
		httpx.WriteError(w, http.StatusConflict, "el email ya está registrado")
		return
	}
	if err != nil {
		slog.Error("error al crear usuario", "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "error interno")
		return
	}

	// Responder 201 con el usuario creado (ya trae id y fechas)
	httpx.WriteJSON(w, http.StatusCreated, usuario)
}

// Crer maneja el PUT /api/usuarios/{id} que se encarga de actualizar un usuario
// el admin puede actualizar cualquiera, el usuario solo el suyo sin cambiar el rol
func (h *UsuarioHandler) Actualizar(w http.ResponseWriter, r *http.Request) {
	// creamos una request para actualizar el usuario
	// no se hace var usuario Models.Usuario porque asi no controlamos lo que manda el usuario, aqui si
	// asi evitamos el mass assigment
	var req actualizarUsuarioRequest

	//a quien quieres editar? quien lo pide? tiene permiso?
	id, err := leerID(r)
	if err != nil {
		// si leerID falla devolvemos 400
		httpx.WriteError(w, http.StatusBadRequest, "el id no es válido")
		return //return para que si un usuario no autenticado intenta algo que no puede el return lo saque
	}

	claims, ok := auth.ClaimsDesdeContexto(r.Context())

	if !ok {
		//si el usuario no esta autorizado usamos el 401
		httpx.WriteError(w, http.StatusUnauthorized, "no autenticado")
		return
	}

	if !puedeAcceder(claims, id) {
		//si el usuario no tiene permisos devolvemos 403
		httpx.WriteError(w, http.StatusForbidden, "no tienes permisos para esta acción")
		return
	}

	// leer y validar el json
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "cuerpo de la peticion invalido")
		return
	}
	err = req.validar()
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "datos del usuario invalidos")
		return
	}

	//cargamos el usuario actual de la base de datos
	usuario, err := h.repo.ObtenerPorID(r.Context(), id)

	if errors.Is(err, repository.ErrNoEncontrado) {
		//si el usuario no existe devolvemos 404
		httpx.WriteError(w, http.StatusNotFound, "usuario no encontrado")
		return
	}
	if err != nil {
		slog.Error("error al obtener usuario para actualizar", "id", id, "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "error interno")
		return
	}

	// 4. Reglas del rol (solo si viene un rol distinto al actual)
	if req.Rol != "" && req.Rol != usuario.Rol {
		if claims.Rol != models.RolAdmin {
			httpx.WriteError(w, http.StatusForbidden, "no puedes cambiar tu propio rol")
			return
		}

		idToken, _ := claims.UsuarioID()
		if idToken == id {
			httpx.WriteError(w, http.StatusBadRequest, "no puedes cambiar tu propio rol")
			return
		}
	}

	//Copiar los campos del dto sobre el usuario actual
	usuario.Nombre = req.Nombre
	usuario.Apellido = req.Apellido
	usuario.Email = req.Email
	usuario.Telefono = req.Telefono
	if req.Rol != "" {
		usuario.Rol = req.Rol // el rol solo si vino
	}
	// Guardar los cambios (404 y 409)
	err = h.repo.Actualizar(r.Context(), usuario)
	if errors.Is(err, repository.ErrNoEncontrado) {
		httpx.WriteError(w, http.StatusNotFound, "usuario no encontrado")
		return
	}
	if errors.Is(err, repository.ErrEmailDuplicado) {
		httpx.WriteError(w, http.StatusConflict, "el email ya está registrado")
		return
	}
	if err != nil {
		slog.Error("error al actualizar usuario", "id", id, "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "error interno")
		return
	}

	// Si vino una contraseña nueva: hashear y guardar
	if req.Password != "" {
		hash, err := auth.HashPassword(req.Password)
		if err != nil {
			slog.Error("error al hashear contraseña", "error", err)
			httpx.WriteError(w, http.StatusInternalServerError, "error interno")
			return
		}
		if err := h.repo.CambiarPassword(r.Context(), id, hash); err != nil {
			slog.Error("error al cambiar contraseña", "id", id, "error", err)
			httpx.WriteError(w, http.StatusInternalServerError, "error interno")
			return
		}
	}

	// Responder con el usuario actualizado
	httpx.WriteJSON(w, http.StatusOK, usuario)

}

// Eliminar maneja DELETE /api/usuarios/{id}: elimina un usuario y, por el
// ON DELETE CASCADE de la base de datos, también todas sus tareas.
// Solo para admins (RequireAdmin); un admin no puede eliminarse a sí mismo.
func (h *UsuarioHandler) Eliminar(w http.ResponseWriter, r *http.Request) {
	// 1. ¿A quién quieren eliminar?
	id, err := leerID(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "id inválido")
		return
	}

	// 2. ¿Quién lo pide?
	claims, ok := auth.ClaimsDesdeContexto(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "no autenticado")
		return
	}

	// 3. Un admin no puede eliminarse a sí mismo
	idToken, err := claims.UsuarioID()
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "token inválido")
		return
	}
	if idToken == id {
		httpx.WriteError(w, http.StatusBadRequest, "no puedes eliminar tu propia cuenta")
		return
	}

	// 4. Eliminar
	err = h.repo.Eliminar(r.Context(), id)
	if errors.Is(err, repository.ErrNoEncontrado) {
		httpx.WriteError(w, http.StatusNotFound, "usuario no encontrado")
		return
	}
	if err != nil {
		slog.Error("error al eliminar usuario", "id", id, "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "error interno")
		return
	}

	// 5. Responder 204 sin cuerpo
	w.WriteHeader(http.StatusNoContent)
}
