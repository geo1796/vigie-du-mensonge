# Vigie du mensonge - project guidelines

This document provides essential guidance for contributing to the project.

## Project overview

Vigie du mensonge is a website aiming to reference lies and falsehoods of the French governments.

## Frontend

### Frontend Tech stack

- Vite - React - TypeScript
- Tanstack (Router, Store)
- ky
- Tailwind CSS
- Shadcn

### Frontend Structure

- clean separation of concerns
- generated routes from tanstack router placed in `/frontend/vigie-du-mensonge/src/routes`
- generated code and components from shadcn placed in `/frontend/vigie-du-mensonge/src/core/shadcn`
- components (other than shadcn) placed in `/frontend/vigie-du-mensonge/src/core/components`
- dependencies placed in `/frontend/vigie-du-mensonge/src/core/dependencies`
- models placed in `/frontend/vigie-du-mensonge/src/core/models`
- tailwind styles imported in `/frontend/vigie-du-mensonge/src/index.css` - these should be the only styles used

## Backend

The api is fully documented using OpenAPI.

### Backend Tech stack

- Go
- fiber(v2) web framework
- gorm (might be replaced by sqlc in the future for better performance)

### Backend Structure

- clean separation of concerns
- feature-based organization placed in `/backend/api`
- custom libraries such as dependencies, env config, models, and fiber wrapper placed in `/backend/core`

### Backend API components

Each feature is placed in its own package under `/backend/api` and typically contains the following:

- `handler.go`: HTTP handlers/controllers
- `handler_test.go`: handler unit tests
- `service.go`: business logic implementation
- `service_test.go`: service unit tests
- `dto.go`: Data Transfer Objects
- `repository.go`: optional feature-specific database access layer
- `integration_test.go`: (MANDATORY) one integration test for successful request/response cycle, using testcontainers
  for production like database and repositories

## Database

PostgresSQL 17

schema.sql placed in `/database/schema.sql`


