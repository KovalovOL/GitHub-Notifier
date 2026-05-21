-include .env

.PHONY: migrate-create migrate-up migrate-down migrate-fix run-sql

# Database migrations
DB_URL=postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

migrate-up:
	migrate -path migrations -database "$(DB_URL)" -verbose up

migrate-down:
	migrate -path migrations -database "$(DB_URL)" -verbose down 1

migrate-fix:
	migrate -path migrations -database "$(DB_URL)" force $(version)

migrate-version:
	migrate -path migrations -database "$(DB_URL)" version

migrate-force:
	migrate -path migrations -database "$(DB_URL)" force $(v)


# Db in docker 
CONTAINER_NAME=postgres_db

run-sql:
	docker exec -it $(CONTAINER_NAME) psql -U $(DB_USER) -d $(DB_NAME) 