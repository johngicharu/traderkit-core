build-server:
	go build -o ./bin/server ./cmd/server

build-controller:
	go build -o ./bin/controller ./cmd/controller

dev:
	wgo -file .go -file .env clear :: wgo run ./cmd/server/main.go :: wgo run ./cmd/controller/main.go
