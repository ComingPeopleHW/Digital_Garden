GO ?= go
GOCACHE ?= $(CURDIR)/backend/.cache/go-build
GOTMPDIR ?= $(CURDIR)/backend/tmp

.PHONY: backend-test backend-build backend-run frontend-build

backend-test:
	cd backend && GOTOOLCHAIN=local GOCACHE=$(GOCACHE) $(GO) test ./...

backend-build:
	mkdir -p backend/bin backend/tmp backend/.cache/go-build
	cd backend && GOTOOLCHAIN=local GOCACHE=$(GOCACHE) GOTMPDIR=$(GOTMPDIR) CGO_ENABLED=1 $(GO) build -ldflags=-linkmode=external -o bin/server ./cmd/server

backend-run: backend-build
	cd backend && ./bin/server

frontend-build:
	cd frontend && pnpm build
