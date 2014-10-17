package main

import (
	"fmt"
	"log"
	"os"

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

	channelName := os.Args[1]

	log.Println("Opening channel ", channelName)
	q, err := ch.QueueDeclare(
		channelName,
		false,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Could not declare Queue")

	body := os.Args[2]
	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	failOnError(err, "Did not send correctly")
}
