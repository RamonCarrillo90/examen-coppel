# Gestión de Usuarios y Tareas

Aplicación web full-stack para administrar usuarios y las tareas asignadas a cada uno, con inicio de sesión y roles.

> 🚧 En construcción. Este README se completará con arquitectura, decisiones técnicas, capturas y credenciales de prueba.

## Stack

| Capa | Tecnología |
|---|---|
| Frontend | Angular 21+ |
| Reverse proxy | Nginx |
| Backend | Go (librería estándar `net/http`) |
| Base de datos | PostgreSQL 17 |
| Migraciones | golang-migrate |
| Contenedores | Docker + Docker Compose |

## Cómo ejecutarlo

Requisitos: Docker Desktop (o Docker Engine + Compose v2).

```bash
cp .env.example .env        # ajusta las contraseñas si quieres
docker compose up --build
```

| Servicio | URL |
|---|---|
| API | http://localhost:8080/api/health |
| PostgreSQL | `localhost:5432` (credenciales en `.env`) |

## Estructura

```
.
├── docker-compose.yml        # Orquestación de servicios
├── .env.example              # Plantilla de configuración
├── Makefile                  # Atajos (make up, make logs, make psql...)
├── backend/
│   ├── Dockerfile            # Build multi-stage
│   ├── cmd/api/main.go       # Punto de entrada
│   ├── internal/
│   │   ├── config/           # Carga de variables de entorno
│   │   ├── database/         # Conexión a PostgreSQL
│   │   ├── handlers/         # Manejadores HTTP
│   │   └── router/           # Rutas y middlewares
│   └── migrations/           # Esquema versionado de la BD
└── frontend/                 # Angular (en progreso)
```

## Modelo de datos

```
usuarios (1) ──────< (N) tareas
   id                      id
   nombre, apellido        usuario_id  → usuarios.id  ON DELETE CASCADE
   email (único)           titulo, descripcion
   telefono                fecha_limite
   password_hash           estatus: pendiente | en_progreso | completada
   rol: admin | usuario
```
