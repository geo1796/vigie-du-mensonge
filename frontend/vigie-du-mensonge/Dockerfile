# --- Stage 1: Build with Node ---
FROM node:24.8.0-alpine AS builder

# Avoid running as root where possible (builder still needs perms to write node_modules)
WORKDIR /app

# Install only what is needed for build in a clean environment
COPY frontend/vigie-du-mensonge/package.json frontend/vigie-du-mensonge/package-lock.json ./
RUN npm ci --no-audit --no-fund

# Copy the rest of the source
COPY frontend/vigie-du-mensonge/. .

# Build the production assets
RUN npm run build

# --- Stage 2: Build OpenAPI docs ---
FROM node:24.8.0-alpine as openapi-builder

WORKDIR /app

COPY openapi .

RUN npm install -g @redocly/cli
RUN redocly build-docs openapi.yml --o docs.html


# --- Stage 3: Run with unprivileged nginx ---
# nginxinc/nginx-unprivileged listens on 8080 by default and runs as non-root (uid 101)
FROM nginxinc/nginx-unprivileged:1.27-alpine

# Security: ensure a clean image and minimal packages
# (Base image is already minimal; avoid adding extra packages.)

# Copy build output
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy openapi docs
RUN mkdir -p /usr/share/nginx/html/openapi
COPY --from=openapi-builder /app/docs.html /usr/share/nginx/html/openapi/index.html

# Provide hardened nginx configuration (includes SPA fallback and headers)
COPY frontend/vigie-du-mensonge/ops/nginx/default.conf /etc/nginx/conf.d/default.conf

EXPOSE 8080

# Healthcheck hits a cheap endpoint served by nginx
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/healthz || exit 1

# Final user is already non-root (uid 101) in this base image.