COMPOSE	= nerdctl compose 

.PHONY: all clean down load-db up

ifneq (,$(wildcard ./.env))
    include .env
    export
endif

all: up build load-db 
	
build:
	@go mod tidy
	@go build -o bucket ./cli/cmd/bucket 

load-db:
	@sleep 1
	cat init.sql | nerdctl exec -i bucket-db psql -U bucket -d bucket

up: 
	@$(COMPOSE) up --build -d 

down:
	@$(COMPOSE) down 

clean: down 
	-@rm -rf ~/.config/bucket
	-@nerdctl image rm -f bucket-api || true 
	-@nerdctl volume rm -f bucket_bucket-db-data
	-@nerdctl system prune -f  