package main

import (
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
)

type MyMessage struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

func main() {
	consumer, err := sarama.NewConsumer([]string{"kafka:9092"}, nil)
	if err != nil {
		panic(err)
	}
	defer consumer.Close()

	partConsumer, err := consumer.ConsumePartition("ping", 0, sarama.OffsetNewest)
	if err != nil {
		panic(err)
	}
	defer partConsumer.Close()

	for {
		select {
		case msg := <-partConsumer.Messages():
			var receivedMessage MyMessage
			err := json.Unmarshal(msg.Value, &receivedMessage)

			if err != nil {
				log.Printf("Error unmarshaling JSON: %v\n", err)
				continue
			}

			log.Printf("Received message: %+v\n", receivedMessage)

			producer, err := sarama.NewSyncProducer([]string{"kafka:9092"}, nil)
			if err != nil {
				panic(err)
			}
			defer producer.Close()

			resp := &sarama.ProducerMessage{
				Topic: "pong",
				Key:   sarama.StringEncoder(receivedMessage.ID),
				Value: sarama.StringEncoder(receivedMessage.Name + " " + receivedMessage.Value + " ( " + receivedMessage.ID + " ) "),
			}
			producer.SendMessage(resp)
		}
	}
}
