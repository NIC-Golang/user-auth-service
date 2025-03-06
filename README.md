# User authentication service
User authentication and authorization microservice for the e-commerce platform. Handles user registration, login, JWT-based authentication, and role management.

## Requirements:
- Golang 1.21+
- Docker & Docker Compose
- MongoDB (preferably)

### Setting up the environment:

**1. In your cli enter: *git clone https://github.com/username/repository.git***

**2. Enter:  *cd user-auth-service***

**3. Copy .env.example file to your .env:  *cp .env.example .env***

## Available Makefile Commands:
```
make help       Shows the list of commands
make install	Install missing dependencies (go mod tidy)
make run	    Run the server
make stop	    Stop the server
make restart	Restart the server
make compile	Compile the application
make clean	    Clean the cache
```