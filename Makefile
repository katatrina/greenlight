# Define the path to your migrations and the database connection string
MIGRATE_PATH = ./migrations
DB_DSN = $(GREENLIGHT_DB_DSN)

# Define the migration target
migrate-up:
	migrate -path $(MIGRATE_PATH) -database $(DB_DSN) up

migrate-down:
	migrate -path $(MIGRATE_PATH) -database $(DB_DSN) down

server:
	go run ./cmd/api