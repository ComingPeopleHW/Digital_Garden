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

## Production Deployment

For an Ubuntu server with Docker installed:

```bash
git clone https://github.com/ComingPeopleHW/Digital_Garden.git
cd Digital_Garden
cp .env.production.example .env
# edit .env and set POSTGRES_PASSWORD plus FRONTEND_ORIGIN=http://YOUR_SERVER_IP
docker compose -f docker-compose.prod.yml up -d --build
```

The production stack includes:

- Caddy serving the React build on port 80
- Go backend behind `/api/*`
- PostgreSQL with a persistent Docker volume

For IP-only access, Caddy serves plain HTTP. When a domain is available, update `deploy/Caddyfile` from `:80` to the domain name and set `FRONTEND_ORIGIN=https://your-domain`.

## Initial MVP

- Public feed across users
- User profile summaries and public profile pages at `/u/{username}`
- Text and image post cards
- PostgreSQL-backed users, posts, media, and reactions
- Upvote and downvote actions with one reaction per demo user
- Email/username registration, login, logout, and session cookies
- Logged-in profile editing for display name, bio, avatar URL, and location
- Logged-in publishing for text and image URL posts
- API skeleton ready for uploads

## Backend API

```http
GET /healthz
GET /api/me
PATCH /api/me/profile
POST /api/auth/register
POST /api/auth/login
POST /api/auth/logout
GET /api/users
GET /api/users/{username}
GET /api/users/{username}/posts
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
