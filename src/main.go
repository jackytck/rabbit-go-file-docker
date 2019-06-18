package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/streadway/amqp"
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
			err := work(cmd, ch)
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

// merge will cat all then remove all
func merge(c Cmd, ch *amqp.Channel) error {
	dir := c.Args[0]
	outName := c.Args[1]
	doneQueue := c.Done

	parts, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	// cat
	var joined []byte
	var ppath []string
	for _, p := range parts {
		if p.IsDir() {
			continue
		}
		name := p.Name()
		if outName == "" {
			ext := filepath.Ext(name)
			outName = name[:len(name)-len(ext)]
		}
		path := filepath.Join(dir, name)
		ppath = append(ppath, path)
		f, err2 := os.Open(path)
		if err2 != nil {
			return err2
		}
		bytes := make([]byte, p.Size())
		buffer := bufio.NewReader(f)
		n, err2 := buffer.Read(bytes)
		if err2 != nil {
			return err2
		}
		joined = append(joined, bytes[:n]...)
		f.Close()
	}

	// write back
	f, err := os.Create(filepath.Join(dir, outName))
	if err != nil {
		return err
	}
	_, err = f.Write(joined)
	if err != nil {
		return err
	}

	// remove parts
	for _, v := range ppath {
		err2 := os.Remove(v)
		if err2 != nil {
			return err2
		}
	}

	body, _ := json.Marshal(c)
	Publish(ch, doneQueue, body)
	return nil
}

func work(cmd Cmd, ch *amqp.Channel) error {
	var err error
	switch cmd.Ops {
	case "cp":
		// sleep to wait for mkdir to propagate
		s, err2 := strconv.Atoi(os.Getenv("CP_SLEEP"))
		if err2 != nil {
			s = 0
		}
		time.Sleep(time.Duration(s) * time.Second)
		log.Printf("Copying: %q to %q\n", cmd.Args[0], cmd.Args[1])
		_, err = copy(cmd.Args[0], cmd.Args[1])
	case "mv":
		log.Printf("Moving: %q to %q\n", cmd.Args[0], cmd.Args[1])
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
		log.Printf("Making directory: %q (%q)\n", cmd.Args[0], cmd.Args[1])
		syscall.Umask(0)
		m, _ := strconv.ParseInt(cmd.Args[1], 0, 32)
		err = os.MkdirAll(cmd.Args[0], os.FileMode(m))
	case "merge":
		// e.g. {"ops":"merge","args":["/tmp/multi-bunny","bunny.zip"],"id":"12345","done":"merge-done"}
		log.Printf("Merging: %q to %q", cmd.Args[0], cmd.Args[1])
		err = merge(cmd, ch)

	}
	return err
}
