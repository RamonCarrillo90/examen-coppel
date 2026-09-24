package router

import (
	"net/http"
	"strings"

	"examen-coppel/backend/internal/auth"
	"examen-coppel/backend/internal/httpx"
	"examen-coppel/backend/internal/models"
)

// RequireAuth exige un JWT válido en el header Authorization ("Bearer <token>").
// Si el token falta o no es válido responde 401; si es válido, guarda los claims
// en el context de la petición y deja pasar al siguiente handler.
func RequireAuth(secreto []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1 y 2. Leer el header y quitarle el prefijo "Bearer "
			token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok || token == "" {
				httpx.WriteError(w, http.StatusUnauthorized, "no autenticado")
				return
			}

			// 3. Verificar firma y expiración
			claims, err := auth.ValidarToken(token, secreto)
			if err != nil {
				httpx.WriteError(w, http.StatusUnauthorized, "token inválido o expirado")
				return
			}

			// 4. Guardar los claims en el context y continuar
			next.ServeHTTP(w, r.WithContext(auth.ConClaims(r.Context(), claims)))
		})
	}
}

// RequireAdmin deja pasar solo a usuarios con rol admin; si no, responde 403.
// Debe usarse DESPUÉS de RequireAuth, porque lee los claims que este guardó.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := auth.ClaimsDesdeContexto(r.Context())
		if !ok {
			httpx.WriteError(w, http.StatusUnauthorized, "no autenticado")
			return
		}
		if claims.Rol != models.RolAdmin {
			httpx.WriteError(w, http.StatusForbidden, "no tienes permiso para esta acción")
			return
		}
		next.ServeHTTP(w, r)
	})
}
