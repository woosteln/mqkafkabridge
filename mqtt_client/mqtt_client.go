package mqtt_client

import (
	"fmt"
	"regexp"
	"strings"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

// MessageType indicates the type of a message
// There are two main message types.
// One for connection / disconnection of devices
// Another indicating messages that have come from devices
type MessageType string

const (
	TYPE_MESSAGE   MessageType = "DEVICE_MESSAGE"
	TYPE_CONNECTED MessageType = "DEVICE_CONNECTED"
)

// Message represents a data payload that either came from a device
// or from the broker
type Message struct {
	Type     MessageType `json:"type"`
	ClientID string      `json:"clientID"`
	Payload  []byte      `json:"payload"`
}

const (
	PATTERN_DEVICE_TOPIC    string = "device_topic"
	PATTERN_EMQ_TOPIC       string = "emq_topic"
	PATTERN_MOSQUITTO_TOPIC string = "mosquitto_topic"
	PATTERN_VERNEMQ_TOPIC   string = "vernemq_topic"
)

var (
	clientPatternDeviceTopic       = regexp.MustCompile("device/(.+)")
	clientPatternEmqSysTopic       = regexp.MustCompile("\\$SYS/brokers/.+/clients/(.+)/.+")
	clientPatternMosquittoSysTopic = regexp.MustCompile("\\$SYS/broker/clients/(.+)")
	clientPatternVerneMQSysTopic   = regexp.MustCompile(`\$SYS/(.+)/mqtt/disconnect|connect/received`)
)

// MQTTClient is a domain specific client for handling messages from particular topics
type MQTTClient struct {
	Client  MQTT.Client
	Handler MessageHandler
}

// MessageHandler is a func type to handle incoming messages received from the broker
type MessageHandler func(string, Message)

// NewMQTTClient returns a new MQTTClient ready to go, but not yet connected
func NewMQTTClient(brokerAddress string, clientID string, user string, password string, cleansession bool) MQTTClient {

	opts := MQTT.NewClientOptions()
	opts.AddBroker(brokerAddress)
	opts.SetClientID(clientID)
	opts.SetUsername(user)
	opts.SetPassword(password)
	opts.SetCleanSession(cleansession)
	opts.SetAutoReconnect(true)
	fmt.Printf("Creating client with broker %s, clientid %s\n", brokerAddress, clientID)
	cli := MQTT.NewClient(opts)
	return MQTTClient{Client: cli}

}

// Connect connects the underlying MQTT client to the broker
func (m *MQTTClient) Connect() {
	cli := m.Client
	if token := cli.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	fmt.Println("Connected")
}

// Subscribe subscribes the underlying MQTT client to relevant $SYS and
// device/+ topics on the mqtt broker, and sets up message handling for
// any messages it receives
func (m *MQTTClient) Subscribe() {

	fmt.Println("Subscribing")

	deviceTopic := "device/+"
	deviceQos := 1
	deviceHandler := MQTT.MessageHandler(func(c MQTT.Client, msg MQTT.Message) {
		clientID := parseClientIdFromTopic(msg.Topic(), PATTERN_DEVICE_TOPIC)
		m.handleMessage(clientID, msg.Payload(), TYPE_MESSAGE)
	})

	if token := m.Client.Subscribe(deviceTopic, byte(deviceQos), deviceHandler); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	// EMQtt topic
	connectEmqTopic := "$SYS/brokers/+/clients/+/+"
	// Mosquitto topic
	connectMosquittoTopic := "$SYS/broker/clients/+"
	// VerneMQ topics
	connectVerneTopic := "$SYS/+/mqtt/connect/received"
	disconnectVerneTopic := "$SYS/+/mqtt/disconnect/received"

	connectQos := 2
	connectEmqHandler := MQTT.MessageHandler(func(c MQTT.Client, msg MQTT.Message) {
		clientID := parseClientIdFromTopic(msg.Topic(), PATTERN_EMQ_TOPIC)
		var js string
		if strings.Contains(msg.Topic(), "disconnected") {
			js = fmt.Sprintf(`{"clientID":%s,"connected":false}`, clientID)
		} else {
			js = fmt.Sprintf(`{"clientID":%s,"connected":true}`, clientID)
		}
		m.handleMessage(clientID, []byte(js), TYPE_CONNECTED)
	})
	connectMosquittoHandler := MQTT.MessageHandler(func(c MQTT.Client, msg MQTT.Message) {
		clientID := parseClientIdFromTopic(msg.Topic(), PATTERN_MOSQUITTO_TOPIC)
		m.handleMessage(clientID, msg.Payload(), TYPE_CONNECTED)
	})
	connectVerneHandler := MQTT.MessageHandler(func(c MQTT.Client, msg MQTT.Message) {
		clientID := parseClientIdFromTopic(msg.Topic(), PATTERN_VERNEMQ_TOPIC)
		var js string
		if strings.Contains(msg.Topic(), "disconnect") {
			js = fmt.Sprintf(`{"clientID":%s,"connected":false}`, clientID)
		} else {
			js = fmt.Sprintf(`{"clientID":%s,"connected":true}`, clientID)
		}
		m.handleMessage(clientID, []byte(js), TYPE_CONNECTED)
	})

	if token := m.Client.Subscribe(connectEmqTopic, byte(connectQos), connectEmqHandler); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	if token := m.Client.Subscribe(connectMosquittoTopic, byte(connectQos), connectMosquittoHandler); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	if token := m.Client.Subscribe(connectVerneTopic, byte(connectQos), connectVerneHandler); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	if token := m.Client.Subscribe(disconnectVerneTopic, byte(connectQos), connectVerneHandler); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

}

func (m *MQTTClient) handleMessage(clientID string, message []byte, typ MessageType) {
	msg := Message{
		ClientID: clientID,
		Payload:  message,
		Type:     typ,
	}
	if m.Handler != nil {
		m.Handler(clientID, msg)
	}
}

// HandleMessage registers a message handler function to the client
func (m *MQTTClient) HandleMessage(fn func(clientID string, msg Message)) {
	m.Handler = MessageHandler(fn)
}

// Disconnect disconnects the underlying MQTT client
func (m *MQTTClient) Disconnect() {
	m.Client.Disconnect(240)
}

func parseClientIdFromTopic(topic string, pattern string) string {
	var regex regexp.Regexp
	switch pattern {
	case PATTERN_DEVICE_TOPIC:
		regex = *clientPatternDeviceTopic
		break
	case PATTERN_EMQ_TOPIC:
		regex = *clientPatternEmqSysTopic
		break
	case PATTERN_MOSQUITTO_TOPIC:
		regex = *clientPatternMosquittoSysTopic
		break
	case PATTERN_VERNEMQ_TOPIC:
		regex = *clientPatternVerneMQSysTopic
		break
	}
	if match := regex.FindStringSubmatch(topic); match != nil {
		return match[1]
	}
	return "unknown"
}
