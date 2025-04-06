APP_NAME := iots
CMD_DIR := ./cmd/api
SEED_DIR := ./cmd/seeder

.PHONY: run build tidy test start sleep seed

run:
	go run $(CMD_DIR)/main.go
build:
	go build -o bin/$(APP_NAME) $(CMD_DIR)/main.go
tidy:
	go mod tidy
test:
	go test ./...
# lint:
# 	golangci-lint run ./...
docker-up:
	sudo docker compose up -d
docker-down:
	sudo docker compose down
sleep:
	sleep 1
seed:
	go run $(SEED_DIR)/main.go

start: docker-up sleep run