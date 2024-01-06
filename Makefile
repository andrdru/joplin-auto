DOCKER_COMPOSE_CMD="docker compose --file docker/docker-compose.yaml"

# up start local env
.PHONY: up
up: down
	@ "${DOCKER_COMPOSE_CMD}" up -d --build

# down stop local env
.PHONY: down
down:
	@ "${DOCKER_COMPOSE_CMD}" down --remove-orphans
