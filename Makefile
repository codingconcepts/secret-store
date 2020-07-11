.PHONY: client, server

client:
	go run ./client/cmd/main.go

server:
	REDIS_ADDR="localhost:6379" REDIS_PASS="" go run ./server/cmd/main.go

redis:
	docker run -d -p 6379:6379 --name redis redis