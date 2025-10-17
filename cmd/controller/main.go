package main

import (
	"backend/internal/common"
	"backend/internal/controller"
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
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

	serverWsURL := os.Getenv("SERVER_WS_URL")
	if serverWsURL == "" {
		log.Fatal("SERVER_WS_URL not set")
	}

	controllerId := os.Getenv("CONTROLLER_ID")
	if controllerId == "" {
		log.Fatal("CONTROLLER_ID not set")
	}

	apiToken := os.Getenv("CONTROLLER_API_TOKEN")
	if apiToken == "" {
		log.Fatal("CONTROLLER_API_TOKEN not set")
	}

	controllerCapacity := os.Getenv("CONTROLLER_CAPACITY")
	if controllerCapacity == "" {
		log.Fatal("CONTROLLER_CAPACITY not set")
	}

	// to int
	intCtrlCapacity, err := strconv.Atoi(controllerCapacity)
	if err != nil {
		log.Fatal("Invalid CONTROLLER_CAPACITY set: %s - int required", controllerCapacity)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srvConfig := common.ControllerConfig{
		Id:          controllerId,
		Token:       apiToken,
		ServerWsUrl: serverWsURL,
		Capacity:    intCtrlCapacity,
	}

	ctrl, err := controller.NewController(ctx, srvConfig)

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

	log.Println("\nController Shutdown signal received...bye")
}
