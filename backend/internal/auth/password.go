// auth es el paquete que verificara que las contraseñas esten bien hasheadas
package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashea la contraseña convirtiendola en byte
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return " ", fmt.Errorf("hashear password: %w", err)
	}
	return string(hash), err
}

// VerificarPassword indica si password coincide con el hash bcrypt guardado
func VerificarPassword(hash, password string) bool {
	// Compara en tiempo real, esto hace que no se pueda ir caracter por caracter adivinando mediante el tiempo de respuesta
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil

}
