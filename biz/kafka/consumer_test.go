package kafka

import (
	"log"
	"testing"

	"github.com/Shopify/sarama"
)

func TestConsumer(_ *testing.T) {
	c := NewConsumer([]string{"10.10.10.86:9092"}, "zim-level1", "zim-level1")
	c.Run(func(msg *sarama.ConsumerMessage) {
		log.Println("recv:", msg.Topic, string(msg.Key), string(msg.Value), msg.Partition, msg.Offset)
	})
}
