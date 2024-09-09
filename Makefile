set-env:
	set -o allexport && source .env && set +o allexport
clean-build:
	rm -rf ./data/
	docker compose up -d --build
build-run:
	docker compose up -d --build
test:
	rm -rf ./data/
	COMPOSE_PROFILES=test docker compose up --build -d
	go test ./...
	COMPOSE_PROFILES=test docker compose down -v
start:
	docker compose up -d
rm:
	COMPOSE_PROFILES=test docker compose down -v
grpc:
	chmod +x grpc_gen.sh
	./grpc_gen.sh
