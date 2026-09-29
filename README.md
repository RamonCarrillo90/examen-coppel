# Gestión de Usuarios y Tareas

Aplicación web full-stack para administrar usuarios y sus tareas, con inicio de sesión por roles, tareas disponibles que los usuarios pueden solicitar y un flujo de aprobación para el administrador.

Desarrollada como examen práctico: **Go** + **PostgreSQL** en el backend, **Angular** en el frontend y **Nginx** como reverse proxy, todo orquestado con **Docker Compose**.

---

## Cómo ejecutarlo

**Requisito:** Docker Desktop (o Docker Engine con Compose v2). No hace falta instalar Go, Node ni PostgreSQL.

1. **Copia la configuración de ejemplo:**

   ```bash
   cp .env.example .env
   ```

   En Windows (PowerShell): `Copy-Item .env.example .env`

2. **Completa los valores secretos en `.env`:**

   - `JWT_SECRET`: una cadena aleatoria de al menos 32 caracteres. Para generarla:
     ```bash
     docker run --rm alpine sh -c "head -c 48 /dev/urandom | base64"
     ```
   - `ADMIN_PASSWORD`: la contraseña del administrador inicial (mínimo 8 caracteres).

3. **Levanta todo:**

   ```bash
   docker compose up --build
   ```

4. **Abre http://localhost**

   Si el puerto 80 está ocupado o Windows no permite usarlo (error `ports are not available ... bind: An attempt was made to access a socket in a way forbidden by its access permissions`), agrega `WEB_PORT=8081` a tu `.env`, vuelve a ejecutar `docker compose up` y entra a http://localhost:8081.

### Usuarios de prueba

Al arrancar por primera vez se crean el administrador y datos de ejemplo: 3 usuarios, tareas en distintos estatus y tareas sin asignar, disponibles para solicitarse.

| Rol | Email | Contraseña |
|---|---|---|
| Administrador | el `ADMIN_EMAIL` de tu `.env` (por defecto `admin@gmail.com`) | el `ADMIN_PASSWORD` de tu `.env` |
| Usuario | `ramon@demo.com` | `demo12345` |
| Usuario | `jazmyne@demo.com` | `demo12345` |
| Usuario | `theo@demo.com` | `demo12345` |

Para desactivar los datos de ejemplo, quita `SEED_DEMO: "true"` del servicio `api` en `docker-compose.yml`.

### Comandos útiles

| Comando | Qué hace |
|---|---|
| `docker compose up --build` | Construye y levanta todo |
| `docker compose down` | Detiene los servicios (conserva la base de datos) |
| `docker compose down -v` | Detiene y **borra** la base de datos |
| `docker compose logs -f api` | Logs del backend en vivo |
| `docker compose ps` | Estado de los servicios (solo `web` publica un puerto) |

---

## Funcionalidades

**Administrador**
- Menú de inicio con resumen: usuarios, tareas (pendientes y sin asignar) y solicitudes por revisar.
- Usuarios: listar con buscador (nombre, apellido, email) y filtro por rol; crear, editar, eliminar (**elimina también sus tareas**) y ver el detalle con sus tareas.
- Tareas: listar todas con buscador y filtros (estatus, usuario, sin asignar); crear con o sin usuario; editar; **reasignar** o dejar sin asignar; eliminar.
- Desde el perfil de un usuario, crear tareas para él (el usuario queda fijo y la tarea siempre nace "pendiente").
- Bandeja de **solicitudes**: aprobar o rechazar.

**Usuario**
- Ve y edita su perfil (sin poder cambiar su rol).
- Ve sus tareas y cambia su estatus.
- Ve las **tareas disponibles** (sin asignar) y las **solicita**; el administrador decide. Consulta el estado de sus solicitudes.

**Reglas de negocio**
- Teléfono de 10 dígitos; fecha límite no anterior a hoy (se permite conservar la fecha de una tarea ya vencida al editarla).
- Una tarea nueva siempre inicia como "pendiente".
- Al aprobar una solicitud, la tarea se asigna y las demás solicitudes de esa tarea se rechazan, todo en **una transacción**: una tarea nunca queda asignada dos veces, aunque dos administradores aprueben al mismo tiempo.
- Un administrador no puede eliminarse ni cambiar su propio rol, así el sistema nunca se queda sin administradores.
- Una solicitud rechazada no impide volver a solicitar la misma tarea; solo se evita tener dos solicitudes **pendientes** iguales (índice único parcial).

---

## Arquitectura

```
                  ┌──────────────────────────────────────────────┐
                  │              Docker Compose                  │
                  │                                              │
 Navegador ──────►│  web (Nginx :80)                             │
 http://localhost │   ├─ /        → archivos de Angular          │
                  │   └─ /api/*   → api:8080 ──► db (PostgreSQL) │
                  │                                              │
                  │  migrate: aplica el esquema y termina        │
                  └──────────────────────────────────────────────┘
```

- **Un solo punto de entrada:** cada servicio corre en su propio contenedor, todos en la misma red interna de Docker. Solo Nginx publica un puerto; la API (`api:8080`) y PostgreSQL (`db:5432`) solo son accesibles dentro de esa red.
- **Mismo origen:** el frontend y la API se sirven desde `http://localhost`, así que no hace falta configurar CORS.
- **Orden de arranque garantizado:** `db` (sana) → `migrate` (terminó bien) → `api` (sana) → `web`.

### Backend (Go)

```
backend/
├── cmd/api/main.go          # Arranque: configuración, conexión, rutas, apagado ordenado
├── migrations/              # Esquema versionado (golang-migrate)
└── internal/
    ├── config/              # Variables de entorno
    ├── database/            # Pool de conexiones a PostgreSQL (pgx)
    ├── models/              # Structs de dominio: Usuario, Tarea, Solicitud
    ├── repository/          # Acceso a datos: SQL parametrizado, transacciones
    ├── handlers/            # HTTP: validación, permisos, traducción de errores a códigos
    ├── auth/                # bcrypt, JWT, claims en el contexto
    ├── router/              # Rutas y middlewares (autenticación, rol admin)
    ├── httpx/               # Respuestas JSON y errores con formato uniforme
    └── seed/                # Administrador inicial y datos de ejemplo
```

Capas con una sola responsabilidad: **handler** (HTTP) → **repository** (SQL) → **PostgreSQL**. Los handlers nunca escriben SQL y los repositorios nunca conocen HTTP.

### Frontend (Angular)

```
frontend/src/app/
├── core/
│   ├── models/          # Interfaces que reflejan el JSON de la API
│   ├── services/        # Únicos que hablan con la API (HttpClient)
│   ├── interceptors/    # Agrega el token JWT; cierra sesión ante un 401
│   ├── guards/          # Sesión y rol admin (solo experiencia de uso; la seguridad está en el backend)
│   └── utils/           # Mensajes de error, fechas
├── layout/navbar/
└── pages/               # Una carpeta por pantalla (lazy loading)
```

---

## API REST

Todas las rutas (salvo `health` y `login`) requieren `Authorization: Bearer <token>`.

| Método | Ruta | Quién | Descripción |
|---|---|---|---|
| GET | `/api/health` | público | Estado del servicio |
| POST | `/api/auth/login` | público | Inicia sesión; devuelve token y usuario |
| GET | `/api/auth/me` | sesión | Usuario actual |
| GET | `/api/usuarios?q=&rol=` | admin | Lista con filtros |
| POST | `/api/usuarios` | admin | Crea usuario |
| GET | `/api/usuarios/{id}` | admin o el propio | Usuario **con sus tareas** |
| PUT | `/api/usuarios/{id}` | admin o el propio | Actualiza (el rol solo lo cambia un admin) |
| DELETE | `/api/usuarios/{id}` | admin | Elimina usuario y sus tareas |
| GET | `/api/tareas?q=&estatus=&usuario_id=&sin_asignar=` | admin | Lista con filtros |
| POST | `/api/tareas` | admin | Crea tarea con o sin usuario |
| POST | `/api/usuarios/{id}/tareas` | admin | Crea tarea para un usuario |
| GET | `/api/tareas/disponibles?q=` | sesión | Tareas sin asignar |
| GET | `/api/tareas/{id}` | admin o dueño | Detalle de tarea |
| PUT | `/api/tareas/{id}` | admin | Actualiza, reasigna o desasigna (`usuario_id: null`) |
| PATCH | `/api/tareas/{id}/estatus` | admin o dueño | Cambia solo el estatus |
| DELETE | `/api/tareas/{id}` | admin | Elimina tarea |
| POST | `/api/tareas/{id}/solicitudes` | usuario | Solicita una tarea disponible |
| GET | `/api/solicitudes/mias` | sesión | Solicitudes propias |
| GET | `/api/solicitudes?estatus=` | admin | Bandeja de solicitudes |
| POST | `/api/solicitudes/{id}/aprobar` | admin | Aprueba (transacción) |
| POST | `/api/solicitudes/{id}/rechazar` | admin | Rechaza |

**Errores:** siempre `{"error": "mensaje"}` con el código adecuado: 400 (datos inválidos), 401 (sin sesión), 403 (sin permiso), 404 (no existe), 409 (conflicto: email duplicado, tarea ya asignada, solicitud repetida), 429 (demasiados intentos de login) y 500 (error interno, sin detalles hacia el cliente).

Las peticiones de prueba de todos los endpoints están en [`docs/api.http`](docs/api.http) (extensión REST Client de VS Code).

---

## Decisiones técnicas

| Tecnología | Por qué | Qué se sacrifica |
|---|---|---|
| **Go** (`net/http`, sin framework) | Binario único y ligero; desde Go 1.22 el router estándar soporta métodos y parámetros; cada capa se entiende sin "magia" | Más código que con un framework como Gin |
| **PostgreSQL** | Datos relacionales; la base garantiza reglas: `UNIQUE`, llaves foráneas, `ON DELETE CASCADE`, `CHECK`, índice único parcial y transacciones. Además `RETURNING` e `ILIKE` | — |
| **pgx** (SQL directo) | Control total de cada consulta; acceso a códigos de error de Postgres (23505, 23503) para responder 409/400 | Sin ORM: más SQL escrito a mano |
| **golang-migrate** | Esquema versionado y reproducible | — |
| **JWT + bcrypt** | API sin estado; contraseñas con hash lento y sal | Un JWT no se puede revocar antes de expirar (expiración corta) |
| **Angular** (standalone, signals, zoneless) | Framework completo y estructurado: router, formularios, HttpClient, inyección de dependencias | Más pesado que React o Vue para apps pequeñas |
| **Nginx** | Sirve los estáticos, reverse proxy (mismo origen, sin CORS), gzip, caché y límite de peticiones al login | — |
| **Docker Compose** | El sistema completo con un comando y las mismas versiones en cualquier máquina | Requiere Docker instalado |

**Seguridad**
- Consultas siempre **parametrizadas** (`$1`, `$2`): el texto de los buscadores nunca se concatena al SQL.
- **DTOs** separados de los modelos: el cliente no puede asignar campos que no le corresponden (*mass assignment*), y los campos desconocidos se rechazan.
- Permisos aplicados en el **backend** (middleware y handlers); los guards de Angular solo mejoran la experiencia.
- Una tarea ajena responde **404** en vez de 403, para no revelar que existe.
- El usuario de las solicitudes se toma del **token**, nunca del cuerpo ni de la URL.
- Límite de 5 intentos de login por minuto por IP (Nginx).

---

## Mejoras futuras

- **Tareas grupales** (varios usuarios por tarea): tabla intermedia `tarea_usuarios`, y redefinir qué significa "reasignar" y qué pasa con la tarea al eliminar a uno de sus miembros.
- **Paginación** en las listas, para volúmenes grandes.
- **Refresh tokens** con expiración corta del token de acceso.
- **Notificaciones en tiempo real** de solicitudes (WebSockets o Server-Sent Events).
- **Pruebas automatizadas:** handlers con `httptest`, repositorios contra un PostgreSQL real en Docker y pruebas de extremo a extremo con Playwright, todo en CI con GitHub Actions. Hoy la verificación se hizo con una lista de más de 100 casos manuales, incluidos permisos por API y aprobaciones simultáneas.
- **Endpoint de resumen** (`COUNT`) para el menú de inicio, en lugar de descargar las listas completas para contarlas.
- Cancelar una solicitud propia mientras esté pendiente.
