package auth

import (
	"errors"
	"testing"
	"time"

	"examen-coppel/backend/internal/models"
)

// secretoPrueba es un secreto fijo solo para las pruebas.
var secretoPrueba = []byte("secreto-de-prueba-de-al-menos-32-bytes!!")

// TestGenerarYValidarToken comprueba el caso feliz: un token recién generado
// se valida y conserva el id y el rol del usuario.
func TestGenerarYValidarToken(t *testing.T) {
	token, err := GenerarToken(7, models.RolAdmin, secretoPrueba, time.Hour)
	if err != nil {
		t.Fatalf("GenerarToken devolvió un error: %v", err)
	}

	claims, err := ValidarToken(token, secretoPrueba)
	if err != nil {
		t.Fatalf("ValidarToken rechazó un token válido: %v", err)
	}

	id, err := claims.UsuarioID()
	if err != nil {
		t.Fatalf("UsuarioID devolvió un error: %v", err)
	}
	if id != 7 {
		t.Errorf("id = %d, se esperaba 7", id)
	}
	if claims.Rol != models.RolAdmin {
		t.Errorf("rol = %q, se esperaba %q", claims.Rol, models.RolAdmin)
	}
}

// TestValidarTokenRechazaTokensInvalidos comprueba que ValidarToken rechaza
// todo lo que no sea un token auténtico y vigente.
func TestValidarTokenRechazaTokensInvalidos(t *testing.T) {
	tokenValido, _ := GenerarToken(7, models.RolUsuario, secretoPrueba, time.Hour)
	tokenVencido, _ := GenerarToken(7, models.RolUsuario, secretoPrueba, -time.Minute)
	tokenOtroSecreto, _ := GenerarToken(7, models.RolUsuario,
		[]byte("otro-secreto-distinto-de-32-bytes-o-mas"), time.Hour)

	casos := []struct {
		nombre string
		token  string
	}{
		{"token vencido", tokenVencido},
		{"firmado con otro secreto", tokenOtroSecreto},
		{"token modificado", tokenValido + "x"},
		{"texto que no es un token", "esto-no-es-un-jwt"},
		{"token vacío", ""},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			_, err := ValidarToken(c.token, secretoPrueba)
			if !errors.Is(err, ErrTokenInvalido) {
				t.Errorf("se esperaba ErrTokenInvalido, se obtuvo: %v", err)
			}
		})
	}
}
