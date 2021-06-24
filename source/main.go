package main

import (
	"log"
	"os"

	"github.com/yomorun/yomo-source-mqtt-starter/pkg/utils"

	"github.com/yomorun/y3-codec-golang"
	"github.com/yomorun/yomo-source-mqtt-starter/pkg/receiver"
)

const ShakeDataKey = 0x10

var (
	codec      = y3.NewCodec(ShakeDataKey)
	appName    = getEnvString("SHAKE_SOURCE_APP_NAME", "shake-source")
	zipperAddr = getEnvString("SHAKE_ZIPPER_ADDR", "localhost:9000")
	serverAddr = getEnvString("SHAKE_SOURCE_SERVER_ADDR", "0.0.0.0:1883")
)

type ShakeData struct {
	Topic   string `y3:"0x20"` // Mqtt Topic
	Payload []byte `y3:"0x21"` // Mqtt Payload
	Time    int64  `y3:"0x22"` // Timestamp (ms)
	From    string `y3:"0x23"` // Source IP
}

func main() {
	handler := func(topic string, payload []byte, writer receiver.ISourceWriter) error {
		log.Printf("receive: topic=%v, payload=%v\n", topic, string(payload))

		// Encode data via Y3 codec https://github.com/yomorun/y3-codec.
		data := ShakeData{Topic: topic, Payload: payload, Time: utils.Now(), From: utils.IpAddr()}
		sendingBuf, _ := codec.Marshal(data)

		// send data via QUIC stream.
		_, err := writer.Write(sendingBuf)
		if err != nil {
			log.Printf("stream.Write error: %v, sendingBuf=%#x\n", err, sendingBuf)
			return err
		}

		log.Printf("write: sendingBuf=%#v\n", sendingBuf)
		return nil
	}

	receiver.CreateRunner(appName, zipperAddr).
		WithServerAddr(serverAddr).
		WithHandler(handler).
		Run()
}

func getEnvString(key string, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) != 0 {
		return value
	}
	return defaultValue
}
