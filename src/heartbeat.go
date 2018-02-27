package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"
)

// listenToPing listens to the heart ping queue and return a go channel of that.
func listenToPing(conn *amqp.Connection, ch *amqp.Channel, ping string) <-chan amqp.Delivery {
	DeclareExchange(ch, ping)
	q := DeclareRandQueue(ch)
	err := ch.QueueBind(q.Name, "", ping, false, nil)
	FailOnError(err, "Failed to bind a queue")
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	FailOnError(err, "Failed to register a consumer")
	return msgs
}

// handlePing handles the ping message and return the pong timestamp.
func handlePing(ch *amqp.Channel, msgs <-chan amqp.Delivery, pong string, verbose bool) {
	LoadEnv()
	host := os.Getenv("HOST_NAME")
	mType := os.Getenv("HOST_TYPE")

	onError := func(err error, msg string) bool {
		if err != nil {
			LogRed(msg)
			return true
		}
		return false
	}

	for d := range msgs {
		if verbose {
			LogCyan("Received a Ping:")
			log.Println(string(d.Body))
		}
		mach := Machine{}
		json.Unmarshal(d.Body, &mach)

		// set start time
		mach.Name = host
		mach.Nickname = host
		mach.Type = mType
		mach.Pong = JSONTime(time.Now())

		// publish to pong queue
		data, err := json.Marshal(mach)
		if onError(err, "Failed to marshal json") {
			continue
		}
		err = Publish(ch, pong, data)
		if onError(err, "Failed to publish to "+pong) {
			continue
		}
		machStr, _ := json.MarshalIndent(mach, "", "  ")
		if verbose {
			LogMagenta("\n" + string(machStr))
			LogCyan("Done Pong")
		}
	}
}

// SubscribeHeartbeat subscribes to the 'ping', heartbeat exchange and publishes
// back to 'pong' queue.
// Required to set env variables: HOST_NAME and HOST_TYPE.
func SubscribeHeartbeat(conn *amqp.Connection, ch *amqp.Channel, ping, pong string, verbose bool) {
	pingMsgs := listenToPing(conn, ch, ping)
	go handlePing(ch, pingMsgs, pong, verbose)
}
