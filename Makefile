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
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	export PATH="$PATH:$(go env GOPATH)/bin"
	protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    ./pkg/grpc-api/api.proto
