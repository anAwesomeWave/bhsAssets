#DB_CONTAINER=db  # не нужно, прописал в docker compose зависимоть остальных контейнеров от бд
APP_CONTAINER=app  # docker compose service name
COMPOSE_FILE=docker-compose.yaml
MIGRATION_CONTAINER=migrate
TEST_STAGE=builder
IMAGE_NAME=my_app:latest  # название образа с тестоывым приложением
DB_PATH=db:5432  # SHOULD MATCH WITH docker-compose. NB! other services use internal port


# для разработки. поднимет бд и применит миграции.
dev: migrate-up

# !internal
migrate-build:
	docker compose build $(MIGRATION_CONTAINER)

# !internal
app-build:
	docker compose build $(APP_CONTAINER)

# !internal
test-build:
	# собрать build стадию докерфайла
	docker build --target $(TEST_STAGE) -t $(IMAGE_NAME) -f Dockerfile_App .

# !internal
prod-build: migrate-build app-build
	@echo "Building the applications..."

# Команда для остановки и удаления контейнеров
down:
	@echo "Stopping Docker containers..."

	- docker rmi -f $(IMAGE_NAME) # ошибки здесь возможны, если не было создано тестового образа,
 		# хотя скорее всего проьлем не возникнет, так как в docker compose прописано поле image и билдятся сервисы под этим же именем
	docker compose -f $(COMPOSE_FILE) down -v --remove-orphans


migrate-up: migrate-build
	@echo "Running migrations up..."
	docker compose -f $(COMPOSE_FILE) run $(MIGRATION_CONTAINER) ./migrate -dbPath $(DB_PATH) -up

migrate-test-up: migrate-up # перед тестовыми миграциями должны отработать обычные
	@echo "Running TEST migrations up..."
	docker compose -f $(COMPOSE_FILE) run $(MIGRATION_CONTAINER) ./migrate -dbPath $(DB_PATH) -up -mgPath "./migrations/test"


test: test-build migrate-test-up
	@echo "Running tests..."
	docker compose -f $(COMPOSE_FILE) run $(APP_CONTAINER) go test -v ./... -cover

# не подготавливает бд (докер ее создаст, но не применятся миграции). использзовать Up
# !internal
run-app: app-build
	@echo "Running app..."
	docker compose -f $(COMPOSE_FILE) up $(APP_CONTAINER)

up: migrate-up run-app down
	@echo "Running application with all dependencies.."

.PHONY: up dev test deps build run-all run-db down migrate-up migrate-down run-tests clean
