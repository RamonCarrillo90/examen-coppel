// Package config carga la configuración de la aplicación desde variables de
// entorno, siguiendo el principio "12-factor": el mismo binario sirve para
// desarrollo, pruebas y producción; solo cambia su entorno.
package config

import (
	"fmt"
	"os"
)

// Config agrupa todos los valores configurables del backend.
type Config struct {
	// Port es el puerto HTTP donde escucha la API (por defecto 8080).
	Port string
	// DatabaseURL es la cadena de conexión a PostgreSQL, por ejemplo:
	// postgres://usuario:clave@db:5432/nombre_bd?sslmode=disable
	DatabaseURL string
}

// Load lee las variables de entorno y valida que las obligatorias existan.
// Si falta alguna, devuelve un error claro en lugar de fallar más adelante
// con un mensaje confuso.
func Load() (*Config, error) {
	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("la variable de entorno DATABASE_URL es obligatoria")
	}

	return cfg, nil
}

// getEnv devuelve el valor de la variable de entorno key, o fallback si no
// está definida.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
