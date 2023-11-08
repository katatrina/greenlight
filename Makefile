server:
	go run ./cmd/api

migrateup:
	migrate -source file://migrations -database "%GREENLIGHT_DB_DSN%" -verbose up

.PHONY: server migrateup