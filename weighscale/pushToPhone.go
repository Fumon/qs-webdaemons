package main

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	_ "github.com/lib/pq"
	"github.com/streadway/amqp"
	"github.com/thorduri/pushover"
)

const (
	pushoverAppToken = ""
	userToken        = ""
	responseURLBase  = ""
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s, %s", msg, err)
		panic(fmt.Sprintf("%s, %s", msg, err))
	}
}

func main() {
	log.Println("Life begins")

	// Create a pushover instance
	pushovr, err := pushover.NewPushover(pushoverAppToken, userToken)
	failOnError(err, "Pushover creation error")

	// Connect to DB
	db, err := sql.Open("postgres", "user=appread dbname='quantifiedSelf' sslmode=disable")
	failOnError(err, "Error connecting to database")
	defer db.Close()

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
			// Declare some binary stuff
			log.Println("Recieved: ", m.Body)
			var did int64
			buf := bytes.NewReader(m.Body)
			err := binary.Read(buf, binary.LittleEndian, &did)
			if err != nil {
				log.Println("Problem decoding delivery, ", err)
				panic(fmt.Sprint("Problem decoding delivery, ", err))
			}
			log.Println("Decoded as did ", did)
			// Make the db query
			var timestamp time.Time
			var weight int64
			err = db.QueryRow("SELECT date, weight FROM buffer.weight WHERE did=$1", did).Scan(&timestamp, &weight)
			switch {
			case err == sql.ErrNoRows:
				log.Printf("No rows returned")
			case err != nil:
				log.Fatal("Error while querying db, ", err)
			default:
				log.Println("Got row for did ", did, ", ", timestamp, " - ", weight)
				// Send via pushover with validate url
				msg := pushover.Message{
					Message:  fmt.Sprint("did: ", did, " - ", ((weight / 100.0) * 2.204), " lbs @", timestamp),
					Title:    "New Weigh-in",
					Url:      fmt.Sprint(responseURLBase, "/weighscale/validate/", did),
					UrlTitle: "Validate",
				}
				req, receipt, err = pushovr.Push(msg)
				if err != nil {
					log.Prinln("Problem sending to pushover, ", err)
				}
			}
		}
	}()

	// Die gracefully
	killchan := make(chan os.Signal)
	signal.Notify(killchan, os.Interrupt, os.Kill)

	log.Print(" [*] Waiting on messages")

	<-killchan

	log.Println("Shutting down...")
}
