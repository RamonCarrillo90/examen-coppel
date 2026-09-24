package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"examen-coppel/backend/internal/auth"
	"examen-coppel/backend/internal/models"
	"examen-coppel/backend/internal/repository"
)

// duracionToken es cuanto dura una sesion: una jornada laboral.
const duracionToken = 8 * time.Hour

// AuthHandler maneja los endpoints de autenticacion: login y sesion actual
type AuthHandler struct {
	repo      *repository.UsuarioRepository
	jwtSecret []byte
}

// NewAuthHandler crea el handler con el repositorio de usuarios y el secreto para firmar los JWT.
func NewAuthHandler(repo *repository.UsuarioRepository, jwtSecret []byte) *AuthHandler {
	return &AuthHandler{repo: repo, jwtSecret: jwtSecret}
}

// loginRequest es el cuerpo que espera POST /api/auth/login
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// logingResponse es lo que devuelve un login exitoso
type loginResponse struct {
	Token   string          `json:"token"`
	Usuario *models.Usuario `json:"usuario"`
}

// Login maneja POST /api/auth/login: verifica las credenciales y devuelve un JWT.
// Responde 401 con el mismo mensaje si el email no existe o la contraseña es incorrecta.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// 1. Leer el JSON del cuerpo
	var req loginRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
		return
	}

	// 2. Buscar al usuario por email
	usuario, err := h.repo.ObtenerPorEmail(r.Context(), req.Email)

	// 3a. No existe → 401 (mismo mensaje que contraseña incorrecta)
	if errors.Is(err, repository.ErrNoEncontrado) {
		writeError(w, http.StatusUnauthorized, "credenciales inválidas")
		return
	}
	// 4. Cualquier otro error → 500
	if err != nil {
		slog.Error("error al buscar usuario en login", "error", err)
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}

	// 3b. Existe, pero la contraseña no coincide → 401
	if !auth.VerificarPassword(usuario.PasswordHash, req.Password) {
		writeError(w, http.StatusUnauthorized, "credenciales inválidas")
		return
	}

	// 5. Generar el token y responder
	token, err := auth.GenerarToken(usuario.ID, usuario.Rol, h.jwtSecret, duracionToken)
	if err != nil {
		slog.Error("error al generar token", "error", err)
		writeError(w, http.StatusInternalServerError, "error interno")
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{Token: token, Usuario: usuario})
}
