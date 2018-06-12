MQKAFKABRIDGE
=============

A client to attach to an mqtt broker and forward messages on a particular topic match on to
a kafka broker.

This service acts as a mqtt client, and listens to %SYS/ topics. This is not recommended
behaviour, and once a mqtt broker has been selected, best practice is to use or create a 
plugin dedicated to acting as the bridge to kafka.

This client however is designed for testing loads and broker performance between multiple
brokers rather than having to create a plugin for each. See it as a first step in 
assessing you mqtt kafka link.

Supported brokers
-----------------

- EMQtt
- Mosquitto
- VerneMQ

Usage
-----

Run the service and it will attach to the configured mqtt broker as a client.

It will subscribe to the `device/+` topic with a persistent session.

It assumes each downstream client will publish to its own `device/{ID}` topic.

When it receives a mesage on the topic it will wrap it in a kafka message data type 
and publish to kafka.
    - `type` The type of message, `DEVICE_MESSAGE`
    - `payload` The original payload bytes of the message
    - `clientID` The deviceID parsed from the topic

When a client connects or disconnects it will publish a kafka message with
the folllowing properties..
    - `type` The type of message, `DEVICE_CONNECTED`
    - `payload` A connected payload ( see below )
    - `clientID` The clients ID.

The payload will be a message with 2 properties.
    - `clientID` The client's id
    - `connected` true | false

Install & run
-------------

You can run standalone or from a docker container.

You configure the service using command line flags or environment variables.

|      Env      |     Flag      |  Type  |     Default     |                                     Description                                      |
| ------------- | ------------- | ------ | --------------- | ------------------------------------------------------------------------------------ |
| MQTT_BROKER   | mqtt_broker   | string | tcp://mqtt:1883 | MQTT broker address ( note websockets supported if you use ws:// or wss:// protocol) |
| CLIENT_ID     | client_id     | string | mqtttokafka     | Client name on mqtt broker                                                           |
| MQTT_USER     | mqtt_user     | string | ""              | Username for mqtt                                                                    |
| MQTT_PASSWORD | mqtt_password | string | ""              | Password for mqtt                                                                    |
| KAFKA_BROKER  | kafka_broker  | string | kafka:9092      | Kafka broker address                                                                 |
| KAFKA_TOPIC   | kafka_topic   | string | mqkafka         | Kafka topic to publish to                                                            |

### Standalone

#### Go get and install

```
go get -u github.com/woosteln/mqkafkabridge
```

#### Run

```
mqkafkabridge --mqtt_broker=tcp://localhost:1083 --kafka_broker=localhost:9092
```

### Docker

Use the prebuilt docker image.

_cmd line_

```
docker run -d -e MQTT_BROKER=tcp://mqtt:1883 -KAFKA_BROKER=kafka:9092 woosteln/mqkafkabridge:latest
```

_docker compose_

An example stack which gets everything except for an example client up and running.

(See github.com/woosteln/mqclientspawner for a service that will create a load of dummy clients).

```
version: "3"
services:
    
    mqtt: 
        image: emq:latest
        ports:
            - 18083:18083
            - 1883:1883
    
    zookeeper:
        image: wurstmeister/zookeeper
        ports:
            - 2181:2181
    
    kafka:
        image: wurstmeister/kafka
        ports:
            - 9092:9092
        depends_on:
            - zookeeper
        environment:
            KAFKA_ADVERTISED_HOST_NAME: 192.168.99.100
            KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
            KAFKA_AUTO_CREATE_TOPICS_ENABLE: true

    mqkafkabridge:
        image: woosteln/mqkafkabridge
        depends_on:
            - kafka
            - mqtt
        environment:
            - MQTT_BROKER: tcp://mqtt:1883
            - CLIENT_ID: mqkafka
            # Use dashboard user for emq as that has auto access to $SYS topics
            - MQTT_USER: dashboard
            - KAFKA_BROKER: kafka:9092
            - KAFKA_TOPIC: testmqmessages
```