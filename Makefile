.PHONY: help build build-middleware build-external test clean run-middleware run-external docker-build docker-up docker-down

help:
	@echo "Available targets:"
	@echo "  build              - Build all services"
	@echo "  build-middleware   - Build middleware service"
	@echo "  build-external     - Build external endpoint service"
	@echo "  test               - Run tests for all services"
	@echo "  clean              - Clean build artifacts"
	@echo "  run-middleware     - Run middleware service locally"
	@echo "  run-external       - Run external endpoint service locally"
	@echo "  docker-build       - Build Docker images"
	@echo "  docker-up          - Start services with Docker Compose"
	@echo "  docker-down        - Stop services with Docker Compose"

build: build-middleware build-external

build-middleware:
	cd services/middleware && go build -o ../../bin/middleware ./cmd/main.go

build-external:
	cd services/external-endpoint && go build -o ../../bin/external-endpoint ./cmd/main.go

test:
	go test ./pkg/... -v
	cd services/middleware && go test ./... -v
	cd services/external-endpoint && go test ./... -v

clean:
	rm -rf bin/
	rm -f services/middleware/middleware
	rm -f services/external-endpoint/external-endpoint

run-middleware:
	cd services/middleware && go run ./cmd/main.go

run-external:
	cd services/external-endpoint && go run ./cmd/main.go

docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

lint:
	golangci-lint run ./...
