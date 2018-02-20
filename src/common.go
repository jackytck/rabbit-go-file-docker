package main

import "log"

// FailOnError checks common rabbitmq error.
func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
