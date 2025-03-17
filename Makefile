.PHONY: build run dev clean help

# Default target when just running 'make'
all: dev

init:
	@echo "== 👩‍🌾 init =="
	@if [ "$$(uname)" = "Darwin" ]; then \
		echo "Installing dependencies for macOS..."; \
		brew install go node pre-commit golangci-lint; \
		brew upgrade golangci-lint; \
	elif [ "$$(uname)" = "Linux" ]; then \
		echo "Installing dependencies for Linux..."; \
		echo "Please install these packages using your distribution's package manager:"; \
		echo "- go (golang)"; \
		echo "- nodejs"; \
		echo "- pre-commit"; \
		echo "- golangci-lint"; \
		echo "For Ubuntu/Debian: sudo apt-get install golang nodejs"; \
		echo "For Fedora: sudo dnf install golang nodejs"; \
		echo "For golangci-lint: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin"; \
		echo "For pre-commit: pip install pre-commit"; \
	else \
		echo "Installing dependencies for Windows..."; \
		echo "Please install these packages:"; \
		echo "- Go from https://golang.org/dl/"; \
		echo "- Node.js from https://nodejs.org/"; \
		echo "- pre-commit using pip: pip install pre-commit"; \
		echo "- golangci-lint from https://golangci-lint.run/usage/install/#windows"; \
	fi

	@echo "== pre-commit setup =="
	pre-commit install || echo "Failed to set up pre-commit. Please install it manually."

	@echo "== install ginkgo =="
	go install -mod=mod github.com/onsi/ginkgo/v2/ginkgo
	go get github.com/onsi/gomega/...

	@echo "== install gomock =="
	go install github.com/golang/mock/mockgen@v1.6.0

dev:
	DOCKERFILE=service/Dockerfile \
    docker-compose -f service/docker-compose.yml build && \
    docker-compose -f service/docker-compose.yml up -d && \
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
	docker-compose -f service/docker-compose.yml down
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
