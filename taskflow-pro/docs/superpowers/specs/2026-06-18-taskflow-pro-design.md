# TaskFlow Pro Design

## Goal

TaskFlow Pro is a full-stack team task collaboration system built for a Go backend portfolio. It demonstrates Gin routing, GORM persistence, MySQL relational modeling, Redis caching and rate limiting, JWT authentication, layered backend architecture, Docker deployment, and a separated frontend presentation layer.

## Product Scope

The application supports a compact but complete workflow:

- Users register and log in with JWT authentication.
- Authenticated users create project spaces.
- Project owners add members by user ID.
- Project members create, update, filter, and move tasks through a kanban-style workflow.
- Users comment on tasks.
- Dashboard APIs return task statistics for the frontend.

## Architecture

The project is split into `backend` and `frontend`.

- `backend`: Go API service using Gin, GORM, MySQL, Redis, JWT, middleware, and service/repository layering.
- `frontend`: independent static HTML/CSS/JS app that calls the backend REST API with `fetch`.
- `docker-compose.yml`: starts MySQL, Redis, backend, and frontend as separate services.

This is intentionally front-end-light. The frontend exists to demonstrate the backend features and prove the API is usable from a real browser.

## Backend Design

Layers:

- `cmd/server`: process entrypoint.
- `internal/config`: environment-based config.
- `internal/database`: MySQL and Redis connection setup.
- `internal/model`: GORM models and API DTOs.
- `internal/repository`: database access.
- `internal/service`: business logic, transactions, password hashing, JWT, Redis cache.
- `internal/handler`: HTTP handlers.
- `internal/middleware`: JWT auth, request logging, CORS, Redis-backed rate limiting.
- `internal/router`: route registration.
- `internal/response`: unified API response helpers.

Primary entities:

- `users`
- `projects`
- `project_members`
- `tasks`
- `comments`

## Redis Usage

Redis is used for concrete backend features:

- API rate limiting by client IP.
- User profile cache after login and `/me`.
- Project dashboard stats cache, invalidated when tasks change.

## API Surface

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `GET /api/v1/auth/me`
- `GET /api/v1/projects`
- `POST /api/v1/projects`
- `POST /api/v1/projects/:id/members`
- `GET /api/v1/projects/:id/stats`
- `GET /api/v1/projects/:id/tasks`
- `POST /api/v1/projects/:id/tasks`
- `PUT /api/v1/tasks/:id`
- `DELETE /api/v1/tasks/:id`
- `GET /api/v1/tasks/:id/comments`
- `POST /api/v1/tasks/:id/comments`
- `GET /health`

## Frontend Design

The frontend is a static separated client:

- Login/register panel.
- Project list and project creation.
- Dashboard metric cards.
- Task filters and kanban columns.
- Task create/edit form.
- Comment list and comment form.

It stores the JWT in `localStorage` and sends `Authorization: Bearer <token>` on API calls. This is intentionally simple so the project remains backend-focused.

## Deployment

Docker Compose provides one-command startup:

- MySQL 8 on host port `3308`.
- Redis 7 on host port `6381`.
- Backend API on host port `8090`.
- Frontend on host port `3001`.

## Testing

Backend tests cover business rules that are useful to explain:

- Password hashing and login validation.
- Task status validation.
- JWT middleware behavior can be manually verified through the browser flow and curl examples.

