# User authentication service
User authentication and authorization microservice for the e-commerce platform. Handles user registration, login, JWT-based authentication, and role management.

## Requirements:
- Golang 1.23+
- Docker & Docker Compose
- MongoDB (preferably)

## Setup Instructions  

Clone the repository and set up the environment:  

```sh
git clone <your-repo-url>
cd user-auth-service
cp .env.example .env
```
## Available Makefile Commands:
```
make help     Shows the list of commands
make install	Install missing dependencies (go mod tidy)
make run	    Run the server
make stop	    Stop the server
make restart	Restart the server
make compile	Compile the application
make clean	  Clean the cache
```

## Running the service:
Make sure Docker Engine is running, then execute:
```
docker compose build
docker compose up
```
To stop the service:
```
docker compose down
```
