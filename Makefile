server:
	go run ./cmd/api

migrateup:
	migrate -source file://migrations -database "%GREENLIGHT_DB_DSN%" -verbose up

migratedown:
	migrate -source file://migrations -database "%GREENLIGHT_DB_DSN%" -verbose down

sqlc:
	docker run --rm -v $(CURDIR):/src -w /src sqlc/sqlc generate

.PHONY: server migrateup migratedown sqlc