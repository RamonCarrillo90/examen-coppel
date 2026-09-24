package auth

import "context"

// claveContexto es un tipo privado: evita choques con claves de otros paquetes.
type claveContexto string

const claveClaims claveContexto = "claims"

// ConClaims devuelve una copia de ctx que lleva los claims del usuario autenticado.
func ConClaims(ctx context.Context, c *Claims) context.Context {
	return context.WithValue(ctx, claveClaims, c)
}

// ClaimsDesdeContexto devuelve los claims que guardó el middleware RequireAuth.
func ClaimsDesdeContexto(ctx context.Context) (*Claims, bool) {
	c, ok := ctx.Value(claveClaims).(*Claims)
	return c, ok
}
