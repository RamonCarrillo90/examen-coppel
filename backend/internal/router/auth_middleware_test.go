package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"examen-coppel/backend/internal/auth"
	"examen-coppel/backend/internal/models"
)

var secretoPrueba = []byte("secreto-de-prueba-de-al-menos-32-bytes!!")

// handlerFinal simula el handler protegido: si la petición llega hasta aquí,
// responde 200. Así sabemos si el middleware la dejó pasar o la bloqueó.
var handlerFinal = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

// hacerPeticion envía una petición GET al handler h con el header Authorization
// indicado (vacío = sin header) y devuelve el código HTTP de la respuesta.
func hacerPeticion(h http.Handler, authorization string) int {
	req := httptest.NewRequest(http.MethodGet, "/api/prueba", nil)
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec.Code
}

// TestRequireAuth comprueba que solo pasan peticiones con un token válido.
func TestRequireAuth(t *testing.T) {
	h := RequireAuth(secretoPrueba)(handlerFinal)

	tokenValido, _ := auth.GenerarToken(1, models.RolUsuario, secretoPrueba, time.Hour)
	tokenVencido, _ := auth.GenerarToken(1, models.RolUsuario, secretoPrueba, -time.Minute)

	casos := []struct {
		nombre        string
		authorization string
		quiero        int
	}{
		{"sin header", "", http.StatusUnauthorized},
		{"sin prefijo Bearer", tokenValido, http.StatusUnauthorized},
		{"token basura", "Bearer abc", http.StatusUnauthorized},
		{"token vencido", "Bearer " + tokenVencido, http.StatusUnauthorized},
		{"token válido", "Bearer " + tokenValido, http.StatusOK},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := hacerPeticion(h, c.authorization); got != c.quiero {
				t.Errorf("status = %d, se esperaba %d", got, c.quiero)
			}
		})
	}
}

// TestRequireAdmin comprueba la cadena completa de una ruta de admin:
// sin sesión → 401, usuario normal → 403, admin → 200.
func TestRequireAdmin(t *testing.T) {
	h := RequireAuth(secretoPrueba)(RequireAdmin(handlerFinal))

	tokenAdmin, _ := auth.GenerarToken(1, models.RolAdmin, secretoPrueba, time.Hour)
	tokenUsuario, _ := auth.GenerarToken(2, models.RolUsuario, secretoPrueba, time.Hour)

	casos := []struct {
		nombre        string
		authorization string
		quiero        int
	}{
		{"sin token", "", http.StatusUnauthorized},
		{"usuario normal", "Bearer " + tokenUsuario, http.StatusForbidden},
		{"admin", "Bearer " + tokenAdmin, http.StatusOK},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := hacerPeticion(h, c.authorization); got != c.quiero {
				t.Errorf("status = %d, se esperaba %d", got, c.quiero)
			}
		})
	}
}

// TestRequireAuthGuardaClaims comprueba que el handler recibe en el context
// los claims del token (id y rol), que es de lo que dependen puedeAcceder y Me.
func TestRequireAuthGuardaClaims(t *testing.T) {
	var recibidos *auth.Claims

	espia := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recibidos, _ = auth.ClaimsDesdeContexto(r.Context())
		w.WriteHeader(http.StatusOK)
	})
	h := RequireAuth(secretoPrueba)(espia)

	token, _ := auth.GenerarToken(42, models.RolAdmin, secretoPrueba, time.Hour)
	hacerPeticion(h, "Bearer "+token)

	if recibidos == nil {
		t.Fatal("el handler no recibió claims en el context")
	}
	id, _ := recibidos.UsuarioID()
	if id != 42 || recibidos.Rol != models.RolAdmin {
		t.Errorf("claims = id %d, rol %q; se esperaba id 42, rol %q", id, recibidos.Rol, models.RolAdmin)
	}
}
