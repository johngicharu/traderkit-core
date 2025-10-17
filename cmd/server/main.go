package main

import (
	"backend/internal/server"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	log.SetOutput(os.Stdout)

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	serverTcpAddr := os.Getenv("SERVER_TCP_ADDR")
	if serverTcpAddr == "" {
		log.Fatal("SERVER_TCP_ADDR not set")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctrl, err := server.NewServer(ctx, serverTcpAddr)

	if err != nil {
		log.Fatalf("encountered error init app: %v", err)
	}

	log.Fatal(ctrl.Run())

	waitForExit()
	time.Sleep(500 * time.Millisecond)
}

func waitForExit() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	<-ch

	log.Println("\nServer Shutdown signal received...bye")
}
