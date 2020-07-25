.PHONY: client server

client:
	go run ./client/cmd/main.go

server_local:
	REDIS_ADDR="localhost:6379" REDIS_PASS="" go run ./server/cmd/main.go

build:
	go build -o app ./main.go

redis:
	docker run -d -p 6379:6379 --name redis redis