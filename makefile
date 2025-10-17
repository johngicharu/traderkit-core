build-server:
	go build -o ./bin/server ./cmd/server

build-controller:
	go build -o ./bin/controller ./cmd/controller
