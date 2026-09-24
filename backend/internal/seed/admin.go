package seed

import (
	"context"
	"errors"
	"examen-coppel/backend/internal/auth"
	"examen-coppel/backend/internal/models"
	"examen-coppel/backend/internal/repository"
	"fmt"
	"log/slog"
)

func CrearAdminSiNoExiste(ctx context.Context, repo *repository.UsuarioRepository, email, password string) error {
	//verificamos que los campos no esten vacios
	if email == "" || password == "" {
		slog.Warn("no se creará el admin inicial: faltan ADMIN_EMAIL o ADMIN_PASSWORD")
		return nil
	}

	_, err := repo.ObtenerPorEmail(ctx, email) // solo nos importa el error, ignoramos el usuario

	if err == nil {
		return nil // si el usuario existe devolvemos null
	}

	if !errors.Is(err, repository.ErrNoEncontrado) {
		return fmt.Errorf("crear admin si no existe: %w", err)
	}

	//hashear la contraseña no se guarda nunca en el texto plano
	hash, err := auth.HashPassword(password)

	if err != nil {
		return fmt.Errorf("hashear la contraseña del admin: %w", err)
	}
	//creamos el modelo del adminsitrador
	admin := &models.Usuario{
		Nombre:       "Administrador",
		Apellido:     "del sistema",
		Email:        email,
		PasswordHash: hash,
		Rol:          models.RolAdmin,
	}

	// si algo falla en el Crear lanzamos el error correspondiente
	if err := repo.Crear(ctx, admin); err != nil {
		return fmt.Errorf("crear admin inicial: %w", err)
	}
	// si todo sale bien lanzamos el mensaje
	slog.Info("admin inicial creado", "email", email)
	return nil
}
