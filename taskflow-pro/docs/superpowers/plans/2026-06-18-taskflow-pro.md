# TaskFlow Pro Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a complete separated frontend/backend Go portfolio project for team task collaboration.

**Architecture:** A Gin API service owns all business logic and data. MySQL stores relational data through GORM. Redis supports rate limiting and cached read models. A static frontend consumes the REST API.

**Tech Stack:** Go, Gin, GORM, MySQL, Redis, JWT, bcrypt, Docker Compose, static HTML/CSS/JavaScript.

---

### Task 1: Backend Foundation

**Files:**
- Create backend Go module, config, database, models, responses, and server entrypoint.

- [ ] Add `backend/go.mod`.
- [ ] Add environment config.
- [ ] Add MySQL/Redis connection setup.
- [ ] Add GORM models and DTOs.
- [ ] Add unified response helpers.
- [ ] Add server entrypoint.
- [ ] Run `go test ./...`.

### Task 2: Auth, Projects, Tasks, Comments

**Files:**
- Create repositories, services, handlers, middleware, and router.

- [ ] Add repositories for users, projects, tasks, and comments.
- [ ] Add auth service with bcrypt and JWT.
- [ ] Add project service with membership checks and stats caching.
- [ ] Add task service with status validation and cache invalidation.
- [ ] Add comment service.
- [ ] Add handlers and route registration.
- [ ] Add middleware for auth, CORS, logging, and Redis rate limiting.
- [ ] Run `go test ./...`.

### Task 3: Frontend

**Files:**
- Create separated static frontend.

- [ ] Add `frontend/index.html`.
- [ ] Add `frontend/styles.css`.
- [ ] Add `frontend/app.js`.
- [ ] Ensure frontend only talks to backend through HTTP APIs.

### Task 4: Deployment and Documentation

**Files:**
- Create Dockerfiles, Compose, README, and MySQL init notes.

- [ ] Add backend Dockerfile.
- [ ] Add frontend Dockerfile and nginx config.
- [ ] Add docker-compose.yml.
- [ ] Add README with run steps and talking points.
- [ ] Run backend build/test verification.

