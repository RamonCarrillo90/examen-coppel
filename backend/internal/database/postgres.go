// Package database se encarga de abrir y configurar la conexión a PostgreSQL.
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect crea un pool de conexiones a PostgreSQL a partir de databaseURL y verifica con un Ping que la base de datos responda.
// Quien la llama es responsable de cerrar el pool con pool.Close().
func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	// 1. Parsear la URL para obtener el objeto de configuración
	cfg, err := pgxpool.ParseConfig(databaseURL)

	if err != nil {
		//Envolver el error con contexto usando fmt.Errorf y %w
		return nil, fmt.Errorf("configuracion de base de datos invalida: %w", err)
	}

	//Ajustamos el tamaño del pool (esto es lo mejor de pgxpool)
	//utilizamos 10 conexiones porque ese numero es mas que suficiente en este caso
	//ademas protegemos a postgres que su numero maximo de conexiones es de 100
	cfg.MaxConns = 10                      //numero maximo de conexiones abiertas
	cfg.MinConns = 2                       //numero minimo de conexiones que siempre estaran abiertas
	cfg.MaxConnIdleTime = 30 * time.Minute // Tiempo maximo que una conexion  puede estar sin usarse antes de cerrarse
	cfg.MaxConnLifetime = time.Hour        //reciclamos conexiones viejas

	//Creamos el pool usando la configuración modificada (usamos NewWithConfig)
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("error al crear el pool de conexiones: %w", err)
	}

	//Hacemos un Ping con limite de tiempo para confirmar que la base de datos realmente responde
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		// Si el Ping falla, cerramos el pool para no dejar conexiones huérfanas
		pool.Close()
		return nil, fmt.Errorf("no se pudo conectar a la base de datos (Ping fallido): %w", err)
	}

	// Si todo sale bien, devolvemos el pool listo para usar
	return pool, nil
}
