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

On this macOS setup, the Go 1.22 toolchain installed by `g` may need external linking:

```bash
GO=/Users/weikang/.g/go/bin/go make backend-run
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

The backend expects `DATABASE_URL` to point at a running PostgreSQL instance. On startup it applies the embedded schema and seed data automatically.

## Initial MVP

- Public feed across users
- User profile summaries
- Text and image post cards
- PostgreSQL-backed users, posts, media, and reactions
- Upvote and downvote actions with one reaction per demo user
- API skeleton ready for auth and uploads

## Backend API

```http
GET /healthz
GET /api/me
POST /api/auth/register
POST /api/auth/login
POST /api/auth/logout
GET /api/users
GET /api/posts
POST /api/posts
POST /api/posts/{postID}/reactions
```

Create post payload:

```json
{
  "title": "A note from my garden",
  "body": "Today I shipped the first version.",
  "imageUrl": "https://example.com/image.jpg"
}
```

Reaction payload:

```json
{
  "type": "upvote"
}
```

Use `X-Demo-User-Key` to simulate a distinct user for reactions before logging in. Once a user is logged in, the session cookie is used.
