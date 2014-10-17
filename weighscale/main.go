package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s, %s", msg, err)
		panic(fmt.Sprintf("%s, %s", msg, err))
	}
}

func main() {
	log.Println("Life begins")

	// Open AMQP
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to amqp")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to create channel, ")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"weights",
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Could not declare Queue")

	// Register consumer
	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to register consumer")

	// Create consumer function
	go func() {
		// Read until channel is closed
		for m := range msgs {
			log.Println("Recieved: ", m.Body)
		}
	}()

	// Die gracefully
	killchan := make(chan os.Signal)
	signal.Notify(killchan, os.Interrupt, os.Kill)

	log.Print(" [*] Waiting on messages")

	<-killchan

	log.Println("Shutting down...")
}
