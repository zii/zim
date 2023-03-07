package kafka

import (
	"time"

	"github.com/Shopify/sarama"
	"zim.cn/base"
)

type Producer struct {
	sp    sarama.SyncProducer
	topic string
}

// 一级生产者
var P1 *Producer

// 二级生产者
var P2 *Producer

func NewProducer(brokers []string, topic string) *Producer {
	c := sarama.NewConfig()
	c.Producer.Return.Successes = true
	c.Producer.Return.Errors = true
	c.Producer.RequiredAcks = sarama.WaitForAll
	c.Producer.Partitioner = sarama.NewHashPartitioner

	sp, err := sarama.NewSyncProducer(brokers, c)
	base.Raise(err)
	p := &Producer{
		sp:    sp,
		topic: topic,
	}
	return p
}

func (p *Producer) SendMessage(key string, value []byte) error {
	m := &sarama.ProducerMessage{}
	m.Topic = p.topic
	m.Key = sarama.StringEncoder(key)
	m.Value = sarama.StringEncoder(value)
	m.Timestamp = time.Now()
	_, _, err := p.sp.SendMessage(m)
	return err
}
