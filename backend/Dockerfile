# Stage 1: Build OpenAPI docs
FROM node:24.8.0-alpine as openapi-builder

WORKDIR /app

COPY openapi/ .

RUN npm install -g @redocly/cli
RUN redocly build-docs openapi.yml --o docs.html

# Stage 2: Build Go application
FROM golang:1.25.0-alpine AS builder

WORKDIR /app

COPY backend/ .

RUN go mod download

RUN go build -o /docker-gs-ping

# Stage 3: Run Go application
FROM golang:1.25.0-alpine

WORKDIR /app

COPY --from=builder /docker-gs-ping /docker-gs-ping

# Copy OpenAPI docs
RUN mkdir -p /app/public
COPY --from=openapi-builder /app/docs.html /app/public/openapi.html

CMD [ "/docker-gs-ping" ]