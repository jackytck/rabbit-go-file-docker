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
)

func main() {
	uri := LoadURI()
	queue := LoadQueue("work")
	wait := LoadWait()
	ping := LoadQueue("ping")
	pong := LoadQueue("pong")

	// connect to rabbit
	conn, ch := ConnectRabbit(uri)
	defer conn.Close()
	defer ch.Close()
	ExitOnClose(conn)
	msgs := ConsumeQueue(conn, ch, queue, 1)

	// subscribe to heartbeater
	SubscribeHeartbeat(conn, ch, ping, pong, false)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			cmd := Cmd{}
			json.Unmarshal(d.Body, &cmd)
			err := work(cmd)
			LogOnError(err, "Failed to work")
			log.Printf("Done")
			d.Ack(false)
			time.Sleep(wait * time.Millisecond)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	forever := make(chan bool)
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
		// sleep to wait for mkdir to propagate
		s, err2 := strconv.Atoi(os.Getenv("CP_SLEEP"))
		if err2 != nil {
			s = 0
		}
		time.Sleep(time.Duration(s) * time.Second)
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
