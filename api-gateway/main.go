package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
)

type MyMessage struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

var responseChannels map[string]chan *sarama.ConsumerMessage
var mu sync.Mutex

func main() {
	responseChannels = make(map[string]chan *sarama.ConsumerMessage)

	producer, err := sarama.NewSyncProducer([]string{"kafka:9092"}, nil)
	if err != nil {
		panic(err)
	}
	defer producer.Close()

	consumer, err := sarama.NewConsumer([]string{"kafka:9092"}, nil)
	if err != nil {
		panic(err)
	}
	defer consumer.Close()

	partConsumer, err := consumer.ConsumePartition("pong", 0, sarama.OffsetNewest)
	if err != nil {
		panic(err)
	}
	defer partConsumer.Close()

	go func() {
		for {
			select {
			case msg := <-partConsumer.Messages():
				responseID := string(msg.Key)
				mu.Lock()
				ch, exists := responseChannels[responseID]
				if exists {
					ch <- msg
					delete(responseChannels, responseID)
				}
				mu.Unlock()
			}
		}
	}()

	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		requestID := fmt.Sprintf("%d", time.Now().UnixNano())

		message := MyMessage{
			ID:    requestID,
			Name:  "Ping",
			Value: "Pong",
		}

		bytes, err := json.Marshal(message)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		msg := &sarama.ProducerMessage{
			Topic: "ping",
			Key:   sarama.StringEncoder(requestID),
			Value: sarama.ByteEncoder(bytes),
		}
		producer.SendMessage(msg)

		responseCh := make(chan *sarama.ConsumerMessage)
		mu.Lock()
		responseChannels[requestID] = responseCh
		mu.Unlock()

		select {
		case responseMsg := <-responseCh:
			c.JSON(200, gin.H{"message": string(responseMsg.Value)})
		case <-time.After(10 * time.Second):
			c.JSON(500, gin.H{"error": "timeout waiting for response"})
		}
	})

	router.Run(":8080")
}
