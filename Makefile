include .env

PROJECTNAME := $(notdir $(CURDIR))
PID := .$(PROJECTNAME).pid
OS := $(shell echo %OS%)

ifeq ($(OS), Windows_NT)
    OS_NAME := Windows
    HELP_CMD := powershell -ExecutionPolicy Bypass -File help.ps1
    DOCKER_BUILD := docker compose build > output.log 2>&1 &
    DOCKER_UP := docker compose up -d
    DOCKER_DOWN := docker compose down
    LOCAL_RUN_CMD := start /B go run cmd/auth/main.go > output.log 2>&1
else
    OS_NAME := Linux
    DOCKER_BUILD := docker-compose build > output.log 2>&1 &
    DOCKER_UP := docker-compose up -d
    DOCKER_DOWN := docker-compose down
    LOCAL_RUN_CMD := go run cmd/auth/main.go > output.log 2>&1 &
endif


ifeq ($(RUN_MODE), docker)
    RUN_CMD := $(DOCKER_UP)
else
    RUN_CMD := $(LOCAL_RUN_CMD)
endif

.PHONY: run
run: ## Run the server
	@echo "Running in $(RUN_MODE) mode on $(OS_NAME) using port:$(AUTH_PORT)..."
	@$(RUN_CMD)

.PHONY: build
build: ## Build the Docker container or prepare local binary
	@echo "Building project..."
	@$(DOCKER_BUILD)

.PHONY: down
down: ## Stop the docker container
	@$(DOCKER_DOWN)

.PHONY:	up
up: ## Run the docker container
	@$(DOCKER_UP)

.PHONY: help
help: ## Show help for each Makefile command
ifeq ($(OS), Windows_NT)
	@$(HELP_CMD)
else
	@egrep -h '\s##\s' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m  %-30s\033[0m %s\n", $$1, $$2}'
endif

.PHONY:	install
install: ## Install missing dependencies
	@echo "Installing dependencies..."
	@go mod tidy

.PHONY: restart
restart: ## Restart the server
ifeq ($(RUN_MODE), docker)
	@$(DOCKER_DOWN)
	@$(DOCKER_UP)
else
	@$(LOCAL_RUN_CMD)
endif

.PHONY: stop
stop: ## Stop the server
	@if exist $(PID) for /f %%i in ('type $(PID)') do taskkill /F /PID %%i
	@if exist $(PID) del /Q $(PID)

.PHONY: compile
compile: ## Compile the Go application
	@echo "Building binary..."
	@go build -o $(PROJECTNAME) cmd/main.go

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning cache..."
	@go clean
