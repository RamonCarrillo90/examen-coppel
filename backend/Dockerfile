# =============================================================================
# Dockerfile multi-stage del backend en Go
# -----------------------------------------------------------------------------
# Etapa 1 (build): imagen grande con el compilador de Go.
# Etapa 2 (final): imagen mínima que solo contiene el binario compilado.
# Resultado: una imagen de ~20 MB en vez de ~800 MB.
# =============================================================================

# ---------- Etapa 1: compilación ----------
FROM golang:1.26-alpine AS build

WORKDIR /src

# Copiamos primero go.mod/go.sum y descargamos dependencias.
# Docker cachea esta capa: si solo cambia el código (y no las dependencias),
# no se vuelven a descargar en cada build.
COPY go.mod go.sum* ./
RUN go mod download

# Ahora sí, el resto del código.
COPY . .

# CGO_ENABLED=0 -> binario estático, no depende de librerías del sistema.
# -ldflags="-s -w" -> quita símbolos de depuración (binario más pequeño).
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/api ./cmd/api

# ---------- Etapa 2: imagen final ----------
FROM alpine:3.22

# wget (incluido en alpine) se usa para el healthcheck de docker-compose.
# Creamos un usuario sin privilegios: el contenedor NO corre como root.
RUN addgroup -S app && adduser -S app -G app

COPY --from=build /bin/api /usr/local/bin/api

USER app
EXPOSE 8080

ENTRYPOINT ["api"]
