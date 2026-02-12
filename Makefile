.PHONY: build run dev test lint migrate-up migrate-down docker-up docker-down clean

APP_NAME := roognis
BUILD_DIR := bin

# ── Build & Run ──────────────────────────────────────────────────────
build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

run: build
	./$(BUILD_DIR)/$(APP_NAME)

dev:
	go run ./cmd/server

# ── Testing ──────────────────────────────────────────────────────────
test:
	go test -race -cover ./...

test-v:
	go test -race -cover -v ./...

# ── Linting ──────────────────────────────────────────────────────────
lint:
	golangci-lint run ./...

fmt:
	gofmt -s -w .
	goimports -w .

# ── Database ─────────────────────────────────────────────────────────
migrate-up:
	migrate -path internal/db/migrations -database "$${DATABASE_URL}" up

migrate-down:
	migrate -path internal/db/migrations -database "$${DATABASE_URL}" down 1

migrate-create:
	migrate create -ext sql -dir internal/db/migrations -seq $(name)

# ── Docker ───────────────────────────────────────────────────────────
docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

# ── Cleanup ──────────────────────────────────────────────────────────
clean:
	rm -rf $(BUILD_DIR)
	go clean -cache -testcache
