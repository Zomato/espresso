.PHONY: build run dev clean help


DOCKER_COMPOSE_CMD := $(shell if command -v docker-compose >/dev/null 2>&1; then echo "docker-compose"; else echo "docker compose"; fi)

# Default target when just running 'make'
all: dev

init:
	@echo "== 👩‍🌾 init =="
	@if command -v brew >/dev/null 2>&1; then \
		echo "Installing with Homebrew..."; \
		brew install pre-commit; \
		brew install golangci-lint; \
		brew upgrade golangci-lint; \
	elif command -v apt-get >/dev/null 2>&1; then \
		echo "Installing with apt..."; \
		sudo apt-get update; \
		sudo apt-get install -y pre-commit; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin; \
	elif command -v dnf >/dev/null 2>&1; then \
		echo "Installing with dnf..."; \
		sudo dnf install -y pre-commit; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin; \
	elif [ "$(shell uname -o 2>/dev/null)" = "Msys" ] || [ "$(shell uname -o 2>/dev/null)" = "Windows_NT" ]; then \
		echo "Windows detected. Please install manually:"; \
		echo "- pre-commit: pip install pre-commit"; \
		echo "- golangci-lint: https://golangci-lint.run/usage/install/#windows"; \
	else \
		echo "Unsupported platform. Please install pre-commit and golangci-lint manually."; \
	fi

	@echo "== pre-commit setup =="
	git config --unset-all --global core.hooksPath 2>/dev/null || true
	pre-commit install
	pre-commit autoupdate
	pre-commit install --install-hooks
	pre-commit install --hook-type commit-msg


dev:
	DOCKERFILE=service/Dockerfile \
    $(DOCKER_COMPOSE_CMD) -f service/docker-compose.yml build && \
    $(DOCKER_COMPOSE_CMD) -f service/docker-compose.yml up -d && \
	open http://localhost:3000

test-lib:
	cd lib && go test -v ./...

test-service: dev
	@echo "Running integration tests..."
	@echo "Waiting for service to be ready..."
	@sleep 5  # Give containers time to start up
	cd service && go test -v ./...

test: test-service test-lib

# Clean up
clean:
	$(DOCKER_COMPOSE_CMD) -f service/docker-compose.yml down
	docker rmi -f docker.io/library/service-espresso-ui docker.io/localstack/localstack docker.io/library/service-espresso


# Show help
help:
	@echo "Available targets:"
	@echo "  make          - build and run the service"
	@echo "  make clean    - Clean up Docker resources"
	@echo "  make test     - Run all tests (both lib and service)"
	@echo "  make test-lib - Run only lib tests"
	@echo "  make test-service - Run only service tests"
	@echo " make init	  - Install dependencies"
