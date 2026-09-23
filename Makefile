# Atajos para los comandos más comunes.
# En Windows: usa Git Bash con make instalado, o copia el comando de la derecha.

.PHONY: up down logs reset psql migrate-new

up:            ## Construye y levanta todos los servicios
	docker compose up --build -d

down:          ## Detiene los servicios (conserva los datos)
	docker compose down

logs:          ## Muestra los logs de la API en vivo
	docker compose logs -f api

reset:         ## Borra TODO (incluida la base de datos) y vuelve a levantar
	docker compose down -v
	docker compose up --build -d

psql:          ## Abre una consola psql dentro del contenedor de Postgres
	docker compose exec db sh -c 'psql -U $$POSTGRES_USER -d $$POSTGRES_DB'

migrate-new:   ## Crea una nueva migración vacía: make migrate-new name=agregar_x
	docker run --rm -v "$(CURDIR)/backend/migrations:/migrations" migrate/migrate:v4.18.3 \
		create -ext sql -dir /migrations -seq $(name)
