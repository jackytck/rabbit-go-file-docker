package main

import (
	"log"
	"strings"

	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
)

// FailOnError checks common rabbitmq error.
func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

// LogOnError checks common file error.
func LogOnError(err error, msg string) {
	if err != nil {
		log.Printf("%s: %s", msg, err)
	}
}

// LoadEnv loads environmental variables from .env
func LoadEnv() {
	err := godotenv.Load()
	FailOnError(err, "Error loading .env file")
}

// ConnectRabbit connects to the rabbit server.
func ConnectRabbit(uri string) (*amqp.Connection, *amqp.Channel) {
	// LogCyan("Connecting:", uri)
	conn, err := amqp.Dial(uri)
	FailOnError(err, "Failed to connect to "+uri)
	ch, err := conn.Channel()
	FailOnError(err, "Failed to open a channel")
	tokens := strings.Split(uri, "@")
	LogBlackOnWhite("Connected", tokens[1])
	return conn, ch
}

// DeclareExchange declares an exchange.
func DeclareExchange(ch *amqp.Channel, name string) {
	err := ch.ExchangeDeclare(
		name,     // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	FailOnError(err, "Failed to declare an exchange "+name)
}

// DeclareQueue declares a queue.
func DeclareQueue(ch *amqp.Channel, queue string) amqp.Queue {
	q, err := ch.QueueDeclare(
		queue, // name
		true,  // durable
		false, // delete when usused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	FailOnError(err, "Failed to declare a queue "+queue)
	return q
}

// DeclareRandQueue creates a random named temporary queue.
func DeclareRandQueue(ch *amqp.Channel) amqp.Queue {
	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when usused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	FailOnError(err, "Failed to declare a temporary queue")
	return q
}

// Publish publishes a message to queue.
func Publish(ch *amqp.Channel, queue string, body []byte) error {
	return ch.Publish(
		"",    // exchange
		queue, // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
}

// PublishExchange publishes a message to the named exchange.
func PublishExchange(ch *amqp.Channel, exchange string, body []byte) error {
	return ch.Publish(
		exchange, // exchange
		"",       // routing key
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
}

// ConsumeQueue connects to a specific queue.
func ConsumeQueue(conn *amqp.Connection, ch *amqp.Channel, queue string, prefetch int) <-chan amqp.Delivery {
	q := DeclareQueue(ch, queue)
	err := ch.Qos(
		prefetch, // prefetch count
		0,        // prefetch size
		false,    // global
	)
	FailOnError(err, "Failed to set QoS")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	FailOnError(err, "Failed to register a consumer")
	return msgs
}
