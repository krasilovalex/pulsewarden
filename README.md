# Pulsewarden

> Self-hosted uptime monitoring platform built with Go, PostgreSQL and a
> production-oriented backend architecture.

Pulsewarden is a monitoring system designed to track availability of
external services and internal applications. Users can create monitors,
configure health checks, collect results, detect incidents and receive
notifications.

The project is being developed as a production-style MVP with a focus on
clean architecture, reliability and operational practices.

------------------------------------------------------------------------

## Project Goals

The main goal of Pulsewarden is to provide a lightweight self-hosted
alternative to commercial uptime monitoring services.

Core capabilities:

-   Create and manage monitors
-   Periodically execute HTTP health checks
-   Store check history
-   Detect service failures
-   Track incidents
-   Send notifications
-   Provide API and web dashboard

------------------------------------------------------------------------

# Architecture Overview

Pulsewarden follows a layered backend architecture.

    cmd/
    ├── api
    ├── worker
    └── migrate

    internal/
    ├── app/
    │   └── api
    │
    ├── domain/
    │   └── monitor
    │
    ├── usecase/
    │   └── monitor
    │
    ├── repository/
    │   └── postgres
    │
    └── platform/
        ├── config
        ├── lifecycle
        ├── logger
        └── postgres

## Components

### API Service

Responsible for:

-   HTTP REST API
-   Request validation
-   Authentication layer (future)
-   Returning monitor information
-   Managing user operations

------------------------------------------------------------------------

### Worker Service

Responsible for background processing:

-   Scheduling checks
-   Executing probes
-   Processing results
-   Detecting failures

------------------------------------------------------------------------

### Migration Service

Responsible for:

-   Database schema migrations
-   Controlled database changes

------------------------------------------------------------------------

# Technology Stack

## Backend

-   Go
-   PostgreSQL
-   pgx / pgxpool
-   Squirrel SQL builder
-   REST API
-   Docker Compose

## Quality Tools

-   go test
-   go test -race
-   go vet

------------------------------------------------------------------------

# Implemented Features

## Infrastructure

Implemented:

-   PostgreSQL integration
-   Connection pooling
-   Environment configuration
-   Structured logging
-   Graceful shutdown
-   Health endpoint
-   Readiness endpoint

------------------------------------------------------------------------

## Database

Implemented:

-   Migration system
-   Monitor database schema
-   PostgreSQL integration tests

------------------------------------------------------------------------

## Domain Layer

Implemented:

Monitor entity with validation rules:

-   Name validation
-   URL validation
-   HTTP method validation
-   Interval validation
-   Timeout validation
-   Expected status validation

------------------------------------------------------------------------

## Repository Layer

Implemented:

-   Create monitor
-   Get monitor by ID
-   PostgreSQL repository abstraction
-   Integration tests with real PostgreSQL

------------------------------------------------------------------------

## API Foundation

Implemented:

-   Request ID middleware
-   Access logging middleware
-   JSON response layer
-   API dependency injection
-   Create monitor endpoint foundation

------------------------------------------------------------------------

# Current Development Status

## Completed

✅ Project architecture\
✅ PostgreSQL infrastructure\
✅ Database migrations\
✅ Monitor domain model\
✅ Repository layer\
✅ Validation layer\
✅ Use case layer\
✅ Integration testing foundation

## In Progress

-   Completing monitor REST API
-   Error response standardization
-   CRUD operations

------------------------------------------------------------------------

# MVP Roadmap

## Phase 1 --- Monitor Management

-   Create monitor
-   Get monitor
-   List monitors
-   Update monitor
-   Delete monitor

------------------------------------------------------------------------

## Phase 2 --- Monitoring Engine

Implementation:

-   Scheduler
-   Worker pool
-   HTTP probe executor
-   Timeout handling
-   Retry strategy
-   Check result storage

------------------------------------------------------------------------

## Phase 3 --- Incident Management

Features:

-   Failure detection
-   Recovery detection
-   Incident lifecycle
-   Incident history

------------------------------------------------------------------------

## Phase 4 --- Notifications

Planned integrations:

-   Telegram
-   Webhooks
-   Email

------------------------------------------------------------------------

## Phase 5 --- Frontend

Planned dashboard:

-   Monitor list
-   Monitor status
-   Create/edit forms
-   Check history
-   Incident timeline

------------------------------------------------------------------------

# Development Principles

The project follows:

-   Explicit dependency injection
-   No global mutable state
-   Separation of domain and infrastructure
-   Repository pattern for database access
-   Context propagation
-   Automated testing
-   Production-oriented practices

------------------------------------------------------------------------

# Running Locally

## Start PostgreSQL

``` bash
docker compose up -d
```

## Run migrations

``` bash
go run ./cmd/migrate -direction up
```

## Start API

``` bash
go run ./cmd/api
```

## Start Worker

``` bash
go run ./cmd/worker
```

------------------------------------------------------------------------

# Testing

Run all tests:

``` bash
go test ./...
```

Race detection:

``` bash
go test -race ./...
```

Static analysis:

``` bash
go vet ./...
```

Integration tests:

``` bash
PULSEWARDEN_INTEGRATION_TESTS=1 \
go test ./internal/repository/postgres -v
```

------------------------------------------------------------------------

# Project Status

Pulsewarden is currently in active MVP development.

The foundation is complete: - backend architecture - database layer -
domain logic - repository layer - testing infrastructure

The next milestones are: - complete monitor CRUD - build monitoring
engine - implement incidents - add notifications - build frontend
dashboard 