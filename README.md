# Go API with PostgreSQL

A clean Go API with PostgreSQL database using Docker.

## Prerequisites

Before running the application, ensure you have the following installed:

    -Docker (Make sure Docker is running)
    -Docker Compose
    -Git

## Quick Start

```bash
# Clone the repository
git clone https://github.com/pycuu/task_for_remitly.git
cd task_for_remitly

# Start the application
docker-compose up --build
```

## Configuration
Environment variables:
- `DB_HOST`: Database host (default: db)
- `DB_PORT`: Database port (default: 5432)
- `DB_USER`: Database user (default: postgres)
- `DB_PASSWORD`: Database password (default: postgres)
- `DB_NAME`: Database name (default: appdb)

## API Documentation
[See API docs here](#) or run the service and visit `http://localhost:8080/docs`
