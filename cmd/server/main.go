package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")

	const url = "amqp://guest:guest@localhost:5672/"
	c, err := amqp.Dial(url)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	fmt.Println("Connection was successful")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	fmt.Println("Shutting down...")
	c.Close()
}
