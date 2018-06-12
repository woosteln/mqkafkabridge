package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/namsral/flag"
	"github.com/woosteln/mqkafkabridge/kafkaproducer"
	"github.com/woosteln/mqkafkabridge/mqtt_client"
)

var (
	mqttBroker   = flag.String("mqtt_broker", "tcp://mqtt:1883", "MQTT broker address")
	mqttClientID = flag.String("client_id", "mqtttokafka", "Client name on mqtt broker")
	mqttUser     = flag.String("mqtt_user", "", "Username for mqtt")
	mqttPassword = flag.String("mqtt_password", "", "Password for mqtt")
	kafka        = flag.String("kafka_broker", "kafka:9092", "Kafka broker address")
	kafkaTopic   = flag.String("kafka_topic", "mqkafka", "Kafka topic to publish to")
)

func main() {

	flag.Parse()

	fmt.Println("Creating kafka")
	kafkaProducer := kafkaproducer.NewKafkaProducer(*kafka, *kafkaTopic)
	fmt.Println("Created kafka")
	cli := mqtt_client.NewMQTTClient(*mqttBroker, *mqttClientID, *mqttUser, *mqttPassword, false)
	cli.Connect()
	cli.Subscribe()
	fmt.Println("Subscribed to mqtt")
	cli.HandleMessage(func(clientID string, message mqtt_client.Message) {
		js, _ := json.Marshal(message)
		kafkaProducer.Publish(clientID, js)
	})
	for {
		time.Sleep(time.Millisecond * 1)
	}

}
