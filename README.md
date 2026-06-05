# Digital Garden

A multi-user homepage and lightweight publishing space for text, images, and simple reactions.

## Stack

- Backend: Go, Chi, PostgreSQL-ready structure
- Frontend: React, Vite, TypeScript
- Local services: Docker Compose for PostgreSQL

## Project Layout

```text
backend/
  cmd/server/        Go HTTP server entrypoint
  internal/api/      API routes and handlers

frontend/
  src/               React app

docker-compose.yml   Local PostgreSQL service
.env.example         Environment variable template
```

## Local Development

Backend:

```bash
cd backend
go mod tidy
go run ./cmd/server
```

Frontend:

```bash
cd frontend
pnpm install
pnpm dev
```

PostgreSQL:

```bash
docker compose up -d postgres
```

## Initial MVP

- Public feed across users
- User profile summaries
- Text and image post cards
- Upvote and downvote actions
- API skeleton ready for persistence, auth, and uploads

