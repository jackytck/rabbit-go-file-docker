package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/streadway/amqp"
)

func main() {
	uri := LoadURI()
	queue := LoadQueue()
	wait := LoadWait()

	conn, err := amqp.Dial(uri)
	FailOnError(err, "Failed to connect to "+uri)
	defer conn.Close()

	ch, err := conn.Channel()
	FailOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		queue, // name
		true,  // durable
		false, // delete when usused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	FailOnError(err, "Failed to declare a queue "+queue)

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
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

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			cmd := Cmd{}
			json.Unmarshal(d.Body, &cmd)
			err = work(cmd)
			LogOnError(err, "Failed to work")
			log.Printf("Done")
			d.Ack(false)
			time.Sleep(wait * time.Second)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func copy(src, dst string) (int64, error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer srcFile.Close()

	srcFileStat, err := srcFile.Stat()
	if err != nil {
		return 0, err
	}

	if !srcFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer dstFile.Close()
	return io.Copy(dstFile, srcFile)
}

func work(cmd Cmd) error {
	var err error
	switch cmd.Ops {
	case "cp":
		log.Printf("Copying: %s to %s\n", cmd.Args[0], cmd.Args[1])
		_, err = copy(cmd.Args[0], cmd.Args[1])
	case "mv":
		log.Printf("Moving: %s to %s\n", cmd.Args[0], cmd.Args[1])
		err = os.Rename(cmd.Args[0], cmd.Args[1])
	case "rm":
		f := cmd.Args[0]
		if strings.HasPrefix(f, "/root") {
			log.Println("Skipped removing:", f)
		} else {
			log.Println("Removing:", f)
			err = os.RemoveAll(f)
		}
	case "mkdir":
		log.Printf("Making directory: %s (%s)\n", cmd.Args[0], cmd.Args[1])
		syscall.Umask(0)
		m, _ := strconv.ParseInt(cmd.Args[1], 0, 32)
		err = os.MkdirAll(cmd.Args[0], os.FileMode(m))
	}
	return err
}
