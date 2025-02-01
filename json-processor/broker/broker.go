package broker

import (
	"log"
	"time"

	"github.com/streadway/amqp"
)

func ListenForMessages(brokerURL string, processFunc func() error) {
	for {
		conn, err := amqp.Dial(brokerURL)
		if err != nil {
			log.Printf("Failed to connect to broker, retrying: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		defer conn.Close()

		channel, err := conn.Channel()
		if err != nil {
			log.Fatalf("Failed to open channel: %v", err)
		}
		defer channel.Close()

		queue := "json-processor-queue"
		_, err = channel.QueueDeclare(queue, true, false, false, false, nil)
		if err != nil {
			log.Fatalf("Failed to declare queue: %v", err)
		}

		msgs, err := channel.Consume(queue, "", true, false, false, false, nil)
		if err != nil {
			log.Fatalf("Failed to consume messages: %v", err)
		}

		log.Printf("Waiting for messages in queue: %s", queue)

		for msg := range msgs {
			log.Printf("Message received: %s", string(msg.Body))
			err := processFunc()
			if err != nil {
				log.Printf("Error processing files: %v", err)
			}
		}
	}
}
