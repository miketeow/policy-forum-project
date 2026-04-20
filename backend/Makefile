# Include variables from the .env file
include .env

export

.PHONY: db-up db-down db-status db-console sqlc-gen

sqlc-gen:
	sqlc generate

db-up:
	goose up

db-down:
	goose down

db-status:
	goose status

db-console:
	psql "$(PGURL)"
