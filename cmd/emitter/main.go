package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	brokerAddr = getEnvString("SHAKE_SOURCE_MQTT_BROKER_ADDR", "tcp://localhost:1883")
	//topic      = getEnvString("SHAKE_SOURCE_MQTT_PUB_TOPIC", "SHAKE")
	interval = getEnvInt("SHAKE_SOURCE_MQTT_PUB_INTERVAL", 500)
	username = getEnvString("SHAKE_SOURCE_MQTT_USERNAME", "yomo")
	password = getEnvString("SHAKE_SOURCE_MQTT_PASSWORD", "yomo")
	//counter    int64
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
		//counter = atomic.AddInt64(&counter, 1)
		//payload := fmt.Sprintf("{\"value\":%v}", counter)
		//pub(client, topic, payload)
		//log.Printf("Publish counter=%d, topic=%v, payload=%v\n", counter, topic, payload)
		//time.Sleep(time.Duration(interval) * time.Millisecond)

		topic := "shake/20210627_cluing/S07"
		payload, _ := json.Marshal(genS07())
		pub(client, topic, payload)
		log.Printf("Publish topic=%v, payload=%s\n", topic, payload)
		time.Sleep(time.Duration(interval) * time.Millisecond)

		topic = "shake/20210627_cluing/S05"
		payload, _ = json.Marshal(genS05())
		pub(client, "shake/20210627_cluing/S05", payload)
		log.Printf("Publish topic=%v, payload=%s\n", topic, payload)
		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
}

type S07 struct {
	TenantId    string  `json:"tenantId"`
	DevEui      string  `json:"devEui"`
	Data        string  `json:"data"`
	CreateDate  int64   `json:"createDate"`
	Temperature float64 `json:"temperature"`
	Vertical    float64 `json:"vertical"`
	Transverse  float64 `json:"transverse"`
}

func genS07() S07 {
	return S07{
		TenantId:    "20210627_cluing",
		DevEui:      "0850533277387820",
		Data:        "CwcMys+69Iks0As4YS4N6A==",
		CreateDate:  1624937248919,
		Temperature: float64(rand.Intn(100)),
		Vertical:    float64(rand.Intn(100)),
		Transverse:  float64(rand.Intn(100)),
	}
}

type S05 struct {
	TenantId   string `json:"tenantId"`
	DevEui     string `json:"devEui"`
	Data       string `json:"data"`
	CreateDate int64  `json:"createDate"`
	Key        string `json:"key"`
}

func genS05() S05 {
	return S05{
		TenantId:   "20210627_cluing",
		DevEui:     "393235307d377504",
		Data:       "AAAQ5gAAAQIIAA==",
		CreateDate: 1624937248919,
		Key:        fmt.Sprintf("0%d00", rand.Intn(10)),
	}
}

func newMqttClient() mqtt.Client {
	options := mqtt.NewClientOptions().
		AddBroker(brokerAddr).
		SetUsername(username).
		SetPassword(password)
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
