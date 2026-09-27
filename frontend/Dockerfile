# =============================================================================
# Imagen del frontend en dos etapas (multi-stage):
#   1. "build": Node instala dependencias y compila Angular.
#   2. "final": Nginx sirve SOLO los archivos compilados.
# La imagen final no lleva Node, ni node_modules, ni el código fuente.
# =============================================================================

# --- Etapa 1: compilar ---
FROM node:24-alpine AS build
WORKDIR /app

# Primero solo package*.json: si no cambian, Docker reutiliza la capa de
# "npm ci" y no vuelve a descargar todas las dependencias en cada build.
COPY package.json package-lock.json ./
RUN npm ci

COPY . .
RUN npx ng build --configuration production

# --- Etapa 2: servir ---
FROM nginx:1.29-alpine

# Nuestra configuración reemplaza la que trae Nginx por defecto.
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY proxy_params_app.conf /etc/nginx/proxy_params_app.conf

# Angular deja la app compilada en dist/<proyecto>/browser.
COPY --from=build /app/dist/frontend/browser /usr/share/nginx/html

EXPOSE 80