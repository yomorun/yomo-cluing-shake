package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/yomorun/y3-codec-golang"

	"github.com/yomorun/yomo/pkg/rx"

	"github.com/yomorun/yomo/pkg/client"
)

const ShakeDataKey = 0x10
const SinkDataKey = 0x11

var (
	appName    = getEnvString("SHAKE_FLOW_APP_NAME", "shake-flow")
	zipperAddr = getEnvString("SHAKE_ZIPPER_ADDR", "localhost:9000")
)

type ShakeData struct {
	Topic   string `y3:"0x20"` // Mqtt Topic
	Payload []byte `y3:"0x21"` // Mqtt Payload
	Time    int64  `y3:"0x22"` // Timestamp (ms)
	From    string `y3:"0x23"` // Source IP
}

func main() {
	ip, port := func() (ip string, port int) {
		ss := strings.Split(zipperAddr, ":")
		ip = ss[0]
		port, _ = strconv.Atoi(ss[1])
		return
	}()

	cli, err := client.NewServerless(appName).Connect(ip, port)
	if err != nil {
		log.Print("❌ Connect to zipper failure: ", err)
		return
	}

	defer cli.Close()
	cli.Pipe(Handler)
}

func Handler(rxStream rx.RxStream) rx.RxStream {
	stream := rxStream.
		Subscribe(ShakeDataKey). // 监听来自shake-source的数据
		OnObserve(decode).       //解码数据到结构体
		Filter(filter).          //只处理关心主题的数据
		Map(homey).              //把数据发送到呼美系统
		Encode(SinkDataKey)      //把处理后的新数据通过新的DataKey给shake-sink监听

	return stream
}

var decode = func(v []byte) (interface{}, error) {
	var mold ShakeData
	err := y3.ToObject(v, &mold)
	if err != nil {
		return nil, err
	}

	rightNow := time.Now().UnixNano() / int64(time.Millisecond)
	fmt.Println(fmt.Sprintf("[%s] %d > topic: %s; payload: %#v ⚡️=%dms", mold.From, mold.Time, mold.Topic, mold.Payload, rightNow-mold.Time))

	return mold, nil
}

var filter = func(i interface{}) bool {
	return i.(ShakeData).Topic == "SHAKE"
}

var homey = func(_ context.Context, i interface{}) (interface{}, error) {
	value := i.(ShakeData)

	// TODO: Send data to HOMEY by https
	fmt.Println(fmt.Sprintf("send to HOMEY with: %v", value))

	return value, nil
}

func getEnvString(key string, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) != 0 {
		return value
	}
	return defaultValue
}
