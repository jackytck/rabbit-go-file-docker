package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

// LoadURI loads the uri of rabbit server from .env.
func LoadURI() string {
	loadEnv()
	user := os.Getenv("RABBIT_USER")
	pwd := os.Getenv("RABBIT_PASSWORD")
	host := os.Getenv("RABBIT_HOST")
	port := os.Getenv("RABBIT_PORT")
	uri := fmt.Sprintf("amqp://%s:%s@%s:%s", user, pwd, host, port)
	return uri
}

// LoadQueue loads the target queue of rabbit server from .env.
func LoadQueue(name string) string {
	loadEnv()
	var queue string
	switch name {
	case "work":
		queue = os.Getenv("RABBIT_QUEUE")
	case "ping":
		queue = os.Getenv("RABBIT_PING")
	case "pong":
		queue = os.Getenv("RABBIT_PONG")
	}

	return queue
}

// LoadWait loads the waiting time in second.
func LoadWait() time.Duration {
	loadEnv()
	wait := os.Getenv("WAIT_TIME")
	t, _ := strconv.ParseInt(wait, 10, 64)
	return time.Duration(t)
}
