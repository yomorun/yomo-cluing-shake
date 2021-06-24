package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	brokerAddr = getEnvString("SHAKE_SOURCE_MQTT_BROKER_ADDR", "tcp://localhost:1883")
	topic      = getEnvString("SHAKE_SOURCE_MQTT_PUB_TOPIC", "SHAKE")
	interval   = getEnvInt("SHAKE_SOURCE_MQTT_PUB_INTERVAL", 500)
	counter    int64
)

func main() {
	client := newMqttClient()
	for {
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			log.Printf("Connect error:%v\n", token.Error())
			time.Sleep(time.Duration(1) * time.Second)
			continue
		}
		break
	}

	for {
		counter = atomic.AddInt64(&counter, 1)

		payload := fmt.Sprintf("{\"value\":%v}", counter)
		pub(client, topic, payload)

		log.Printf("Publish counter=%d, topic=%v, payload=%v\n", counter, topic, payload)
		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
}

func newMqttClient() mqtt.Client {
	options := mqtt.NewClientOptions().
		AddBroker(brokerAddr).
		SetUsername("admin").
		SetPassword("public")
	log.Println("Broker Addresses: ", options.Servers)
	options.SetClientID(fmt.Sprintf("shake-source-pub-%d", time.Now().Unix()))
	options.SetConnectTimeout(time.Duration(0) * time.Second)
	options.SetAutoReconnect(true)
	options.SetKeepAlive(time.Duration(20) * time.Second)
	options.SetMaxReconnectInterval(time.Duration(5) * time.Second)
	options.OnConnect = func(client mqtt.Client) {
		log.Println("Connected")
	}
	options.OnConnectionLost = func(client mqtt.Client, err error) {
		log.Printf("Connect lost: %v", err)
	}
	options.SetOnConnectHandler(func(c mqtt.Client) {
		log.Printf("[client connect state] IsConnected:%v, IsConnectionOpen:%v", c.IsConnected(), c.IsConnectionOpen())
	})
	options.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		log.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	})

	client := mqtt.NewClient(options)

	return client
}

func pub(client mqtt.Client, topic string, payload interface{}) {
	for {
		if token := client.Publish(topic, 1, false, payload); token.Wait() && token.Error() != nil {
			log.Printf("yomo-source Publish error: %s \n", token.Error())
			time.Sleep(time.Duration(interval) * time.Millisecond)
			continue
		}
		break
	}
}

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if len(value) != 0 {
		result, err := strconv.Atoi(value)
		if err != nil {
			return defaultValue
		}

		return result
	}
	return defaultValue
}

func getEnvString(key string, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) != 0 {
		return value
	}
	return defaultValue
}
