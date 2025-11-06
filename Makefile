include .env
export $(shell sed 's/=.*//' .env)

up:
	docker compose --env-file .env up --build

down:
	docker compose down

logs:
	docker compose logs -f

ps:
	docker compose ps

restart:
	docker compose down && docker compose --env-file .env up --build -d

clean:
	docker compose down -v --remove-orphans
	docker system prune -f

migrate:
	docker compose run --rm migrate

test:
	docker compose run --rm api-dev go test ./... -v

dev:
	docker compose up api-dev
