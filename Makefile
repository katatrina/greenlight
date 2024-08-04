# Define the path to your migrations and the database connection string
MIGRATE_PATH = ./migrations
DB_DSN = $(GREENLIGHT_DB_DSN)

# Define the migration target
migrate-up:
	migrate -path $(MIGRATE_PATH) -database $(DB_DSN) -verbose up

migrate-down:
	migrate -path $(MIGRATE_PATH) -database $(DB_DSN) -verbose down

migrate-down1:
	migrate -path $(MIGRATE_PATH) -database $(DB_DSN) -verbose down 1

sqlc:
	sqlc generate

server:
	go run ./cmd/api