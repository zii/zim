package kafka

import (
	"context"
	"time"

	"zim.cn/base/log"

	"zim.cn/base"

	"github.com/Shopify/sarama"
)

type Consumer struct {
	group    sarama.ConsumerGroup
	topic    string
	group_id string
	handler  func(message *sarama.ConsumerMessage) bool
}

func (c *Consumer) Setup(_ sarama.ConsumerGroupSession) error {
	log.Println("setup kafka consumer.", c.group_id)
	return nil
}
func (c *Consumer) Cleanup(_ sarama.ConsumerGroupSession) error {
	log.Println("cleanup kafka consumer.", c.group_id)
	return nil
}
func (c *Consumer) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		if c.handler != nil {
			ok := c.handler(msg)
			if !ok {
				continue
			}
		}
		sess.MarkMessage(msg, "")
	}
	log.Println("close consume.")
	return nil
}

func NewConsumer(brokers []string, topic string, group_id string) *Consumer {
	c := sarama.NewConfig()
	c.Consumer.Offsets.Initial = sarama.OffsetNewest
	c.Consumer.Return.Errors = false
	group, err := sarama.NewConsumerGroup(brokers, group_id, c)
	base.Raise(err)
	cs := &Consumer{
		group:    group,
		topic:    topic,
		group_id: group_id,
	}
	return cs
}

func (c *Consumer) Run(h func(*sarama.ConsumerMessage) bool) {
	c.handler = h
	ctx := context.Background()
	for {
		err := c.group.Consume(ctx, []string{c.topic}, c)
		if err != nil {
			log.Println("group.Consume err:", err)
			time.Sleep(10 * time.Second)
		}
		log.Println("finish.")
	}
}
