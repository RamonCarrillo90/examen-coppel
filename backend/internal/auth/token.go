package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims son los datos que viajan en un JWT
type Claims struct {
	Rol                  string `json:"rol"`
	jwt.RegisteredClaims        // incrusta sub, exp, iat... (composición, no herencia)
}

var ErrTokenInvalido = errors.New("token invalido")

func GenerarToken(usuarioID int, rol string, secreto []byte, duracion time.Duration) (string, error) {
	//guardamos el instante una sola vez asi IssuedAt y ExpiresAt parten donde mismo
	ahora := time.Now()
	claims := Claims{
		Rol: rol, // el dato rol va directo
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: strconv.Itoa(usuarioID), //id del usuario convertido a texto porque el estandar jwt lo define como string
			//usan jwt.NewNumericDate() porque jwt guarda las fechas como segundos
			IssuedAt:  jwt.NewNumericDate(ahora),
			ExpiresAt: jwt.NewNumericDate(ahora.Add(duracion)),
		},
	}
	//arma el token sin firmar indicando el algoritmo HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	//calcula al firma con tu JWT_SECRET y devuelve el texto final xxx.yyy.zzz
	firmado, err := token.SignedString(secreto)
	if err != nil {
		return "", fmt.Errorf("firmar token: %w", err)
	}
	return firmado, nil

}

// ValidarToken verifica la firma y la expiración de tokenString y devuelve sus claims.
// Solo acepta tokens firmados con HS256. Si el token no es válido, el error
// envuelve a ErrTokenInvalido.
func ValidarToken(tokenString string, secreto []byte) (*Claims, error) {
	claims := &Claims{}
	//Verifica la fimra y la expiracion, un token vencido da error sin que hagas nada
	_, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (any, error) { return secreto, nil },

		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTokenInvalido, err)
	}
	return claims, nil
}

// UsuarioID devuelve el id del usuario guardado en el Subject del token.
func (c *Claims) UsuarioID() (int, error) {
	return strconv.Atoi(c.Subject)
}
