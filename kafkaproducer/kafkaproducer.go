package kafkaproducer

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaProducer struct {
	producer kafka.Producer
	topic    string
}

func NewKafkaProducer(kafkaAddr string, topic string) KafkaProducer {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": kafkaAddr})
	if err != nil {
		panic(err)
	}
	kafka := KafkaProducer{
		producer: *p,
		topic:    topic,
	}
	go startReporting(*p)
	return kafka
}

func startReporting(p kafka.Producer) {
	for e := range p.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				fmt.Printf("Delivery failed: %s, %v\n", ev.TopicPartition.Error.Error(), ev.TopicPartition)
			}
		}
	}
}

func (k *KafkaProducer) Publish(id string, message []byte) {
	k.producer.Produce(&kafka.Message{
		Key:            []byte(id),
		TopicPartition: kafka.TopicPartition{Topic: &k.topic, Partition: kafka.PartitionAny},
		Value:          message,
	}, nil)
}
