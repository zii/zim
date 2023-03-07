package main

import (
	"fmt"
	"log"

	"golang.org/x/net/context"

	"github.com/Shopify/sarama"
	"zim.cn/base"
)

var addrs = []string{"10.10.10.86:9092"}

const topic = "zim-level1"
const group_id = "zim-level1"

func test_produce() {
	c := sarama.NewConfig()
	c.Producer.Return.Successes = true
	c.Producer.Return.Errors = true
	c.Producer.RequiredAcks = sarama.WaitForAll
	c.Producer.Partitioner = sarama.NewHashPartitioner

	p, err := sarama.NewSyncProducer(addrs, c)
	base.Raise(err)

	// sendmessage
	m := &sarama.ProducerMessage{}
	m.Topic = topic
	m.Key = sarama.StringEncoder("111")
	m.Value = sarama.StringEncoder("hello")
	par, offset, err := p.SendMessage(m)
	base.Raise(err)
	fmt.Println("send success:", par, offset)
}

type Consumer struct {
}

func (*Consumer) Setup(_ sarama.ConsumerGroupSession) error {
	log.Println("setup consume.")
	return nil
}
func (*Consumer) Cleanup(_ sarama.ConsumerGroupSession) error {
	log.Println("cleanup consume.")
	return nil
}
func (*Consumer) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		log.Println("recv:", msg.Topic, string(msg.Key), string(msg.Value), msg.Partition, msg.Offset)
		//time.Sleep(10 * time.Second)
		//log.Println("done.")
		sess.MarkMessage(msg, "")
	}
	log.Println("close consume.")
	return nil
}

func test_consume() {
	c := sarama.NewConfig()
	c.Consumer.Offsets.Initial = sarama.OffsetNewest
	c.Consumer.Return.Errors = false
	group, err := sarama.NewConsumerGroup(addrs, group_id, c)
	base.Raise(err)
	ctx := context.Background()
	h := &Consumer{}
	for {
		err = group.Consume(ctx, []string{topic}, h)
		base.Raise(err)
		log.Println("finish.")
	}
}

func main() {
	//test_produce()
	test_consume()
}
