include .env
PROJECTNAME=$(notdir $(CURDIR))
PID=.$(PROJECTNAME).pid

## install: Install missing dependencies. Runs `go mod tidy`.
.PHONY: install
install:
	@echo "Installing dependencies..."
	@go mod tidy

## run: Run the server
.PHONY: run
run:
	@echo "Server is running on port:$(AUTH_PORT) ..."
	@go run cmd/auth/main.go > output.log 2>&1 &
	@echo $$! > $(PID)

	

## stop: Stop the server
.PHONY: stop
stop:
	@if exist $(PID) for /f %%i in ('type $(PID)') do taskkill /F /PID %%i
	@if exist $(PID) del /Q $(PID)



## restart: Restart the server
.PHONY: restart
restart: stop run

## compile: Compile the app
.PHONY: compile
compile:
	@echo Building...
	@go build -o $(PROJECTNAME).exe cmd/auth/main.go

## clean: Clean the cache
.PHONY: clean
clean:
	@echo "Cleaning cache..."
	@go clean


## help: Show the commands
.PHONY: help
help:
	@powershell -ExecutionPolicy Bypass -File help.ps1
