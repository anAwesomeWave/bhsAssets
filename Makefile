#DB_CONTAINER=db
APP_CONTAINER=app
COMPOSE_FILE=docker-compose.yaml
#MIGRATION_TOOL=goose
#MIGRATION_DIR=migrations
#APP_BINARY=your_app_binary

# Команда по умолчанию
#.DEFAULT_GOAL := up

# Команда up: запускает контейнеры базы данных и приложения
#up: migrate-up run-all

# Команда dev: запускает только контейнер с базой данных
#dev: run-db

# Команда test: запускает миграции, тесты и откатывает миграции
#test: migrate-up run-tests migrate-down

# Команда для запуска всех контейнеров
#run-all:
#	@echo "Starting all Docker containers..."
#	docker-compose -f $(COMPOSE_FILE) up -d
#
## Команда для запуска только контейнера с базой данных
#run-db:
#	@echo "Starting Docker container for the database..."
#	docker-compose -f $(COMPOSE_FILE) up -d $(DB_CONTAINER)

# Команда для остановки и удаления контейнеров
down:
	@echo "Stopping Docker containers..."
	docker-compose -f $(COMPOSE_FILE) down -v

# Команда для выполнения миграций вверх
migrate-up:
	@echo "Running migrations up..."
	docker compose -f $(COMPOSE_FILE) run migrate ./migrate -dbPath "db:5432" -up
# Команда для выполнения миграций вниз
#migrate-down:
#	@echo "Reverting migrations down..."
#	docker-compose -f $(COMPOSE_FILE) run --rm $(APP_CONTAINER) \
#		$(MIGRATION_TOOL) -dir $(MIGRATION_DIR) -database $(DB_URL) down

# Команда для запуска тестов
run-tests:
	@echo "Running tests..."
	docker compose -f $(COMPOSE_FILE) run --rm $(APP_CONTAINER) go test -v ./...

run-app:
	@echo "Running app..."
	docker compose -f $(COMPOSE_FILE) up $(APP_CONTAINER)

run-all: migrate-up run-app
	@echo "Running all.."
	#docker compose -f $(COMPOSE_FILE) up

# Команда для очистки артефактов
#clean: down
#	@echo "Cleaning up Docker artifacts..."
#	docker system prune -f
#	@echo "Removing binary..."
#	rm -f $(APP_BINARY)

# Дополнительные команды
.PHONY: up dev test deps build run-all run-db down migrate-up migrate-down run-tests clean
