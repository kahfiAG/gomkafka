package main

import (
	"fmt"
	kafka "github.com/Shopify/sarama"
	"github.com/jeffchao/gomkafka/gomkafka"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func run() {
	go handleSignals()
	work()
}

func initConfig() (*gomkafka.KafkaConfig, error) {
	config := &gomkafka.KafkaConfig{}

  if len(os.Args) != 4 {
    printUsage()
    os.Exit(2)
  }

  config.ClientId = os.Args[1]
  for _, h := range strings.Split(os.Args[2], ",") {
    config.Hosts = append(config.Hosts, h)
  }
  config.Topic = os.Args[3]

	if config.ClientId == "" || len(config.Hosts) == 0 || config.Topic == "" {
    printUsage()
		os.Exit(2)
	}

	return config, nil
}

func printUsage() {
  fmt.Printf("Usage: gomkafka id hosts topic\n")
  fmt.Printf("\tid\tKafka client id (REQUIRED)\n")
  fmt.Printf("\thosts\tComma-separated list of host:port pairs (REQUIRED)\n")
  fmt.Printf("\ttopic\tKafka topic (REQUIRED)\n")
}

func work() {
	config, err := initConfig()
	if err != nil {
		panic(err)
	}

	client, producer, err := gomkafka.Gomkafka(config)
	if err != nil {
		panic(err)
	}
	defer client.Close()
	defer producer.Close()

	msg := ""

	for {
		_, err := fmt.Scanf("%s\n", &msg)
		if err != nil {
			return
		}

		err = producer.QueueMessage("monitoring", nil, kafka.StringEncoder(msg))
		if err != nil {
			panic(err)
		}

		select {
		case err = <-producer.Errors():
			if err != nil {
				panic(err)
			}
		default:
			// Perform a noop so sarama can can catch disconnect on the other end.
		}

		time.Sleep(1 * time.Millisecond)
	}
}

func handleSignals() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGUSR1, syscall.SIGTERM, syscall.SIGKILL, os.Interrupt)

	for s := range signals {
		switch s {
		case syscall.SIGUSR1, syscall.SIGTERM, syscall.SIGKILL, os.Interrupt:
			// Catch signals that might terminate the process on behalf all goroutines.
			quit()
		}
	}
}

func quit() {
	// Perform any necessary cleanup here.
	os.Exit(1)
}

func main() {
	run()
}
