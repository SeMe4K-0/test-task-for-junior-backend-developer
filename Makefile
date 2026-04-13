MIGRATIONS_PATH=./migrations

.PHONY: migrate-up
migrate-up:
	migrate -path ${MIGRATIONS_PATH} -database ${DATABASE_DSN} up $(filter-out $@, ${MAKECMDGOALS})

.PHONY: migrate
migrate:
	migrate create -seq -ext sql -dir ${MIGRATIONS_PATH} $(filter-out $@, ${MAKECMDGOALS})

.PHONY: migrate-drop
migrate-drop:
	migrate -path ${MIGRATIONS_PATH} -database ${DATABASE_DSN} drop

.PHONY: seed
seed:
	go run ./cmd/seed/seed.go