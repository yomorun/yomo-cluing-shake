package main

import (
	"crypto/sha1"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/yomorun/yomo-source-mqtt-starter/pkg/utils"

	"github.com/yomorun/y3-codec-golang"
	"github.com/yomorun/yomo-source-mqtt-starter/pkg/receiver"
)

const ShakeDataKey = 0x10

var (
	codec       = y3.NewCodec(ShakeDataKey)
	appName     = getEnvString("SHAKE_SOURCE_APP_NAME", "shake-source")
	zipperAddr  = getEnvString("SHAKE_ZIPPER_ADDR", "localhost:9000")
	serverAddr  = getEnvString("SHAKE_SOURCE_SERVER_ADDR", "0.0.0.0:1883")
	enableDebug = getEnvBool("SHAKE_SOURCE_ENABLE_DEBUG", false)
)

type ShakeData struct {
	Topic   string `y3:"0x20"` // Mqtt Topic
	Payload []byte `y3:"0x21"` // Mqtt Payload
	Time    int64  `y3:"0x22"` // Timestamp (ms)
	From    string `y3:"0x23"` // Source IP
}

func main() {
	handler := func(topic string, payload []byte, writer receiver.ISourceWriter) error {
		//log.Printf("receive: topic=%v, payload=%v\n", topic, string(payload))

		// Encode data via Y3 codec https://github.com/yomorun/y3-codec.
		data := ShakeData{Topic: topic, Payload: payload, Time: utils.Now(), From: utils.IpAddr()}
		sendingBuf, _ := codec.Marshal(data)

		// send data via QUIC stream.
		_, err := writer.Write(sendingBuf)
		if err != nil {
			log.Printf("stream.Write error: %v, sendingBuf=%#x\n", err, sendingBuf)
			return err
		}

		if enableDebug {
			log.Printf("write: topic=%s, payload.hash=%#v, sendingBuf=%#x\n", topic, genSha1(payload), sendingBuf)
		} else {
			log.Printf("write: topic=%s, payload.hash=%#v\n", topic, genSha1(payload))
		}
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

func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if len(value) != 0 {
		flag, err := strconv.ParseBool(value)
		if err != nil {
			return defaultValue
		}
		return flag
	}
	return defaultValue
}

func genSha1(buf []byte) string {
	h := sha1.New()
	h.Write(buf)
	return fmt.Sprintf("%x", h.Sum(nil))
}
