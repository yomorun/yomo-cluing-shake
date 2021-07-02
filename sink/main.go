package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/yomorun/y3-codec-golang"

	"github.com/yomorun/yomo/pkg/rx"

	"github.com/yomorun/yomo/pkg/client"

	"github.com/gin-gonic/gin"

	socketio "github.com/googollee/go-socket.io"
)

const (
	SinkDataKey   = 0x11
	WebSocketRoom = "yomo-demo"
	WebSocketAddr = "0.0.0.0:8000"

	S07Topic = "shake/20210627_cluing/S07"
	S05Topic = "shake/20210627_cluing/S05"
)

var (
	webSocketServer *socketio.Server
	err             error
	appName         = getEnvString("SHAKE_SINK_APP_NAME", "shake-sink")
	zipperAddr      = getEnvString("SHAKE_ZIPPER_ADDR", "localhost:9000")
	enableDebug     = getEnvBool("SHAKE_SINK_ENABLE_DEBUG", false)
)

type SinkData struct {
	Topic   string `y3:"0x20" json:"topic"`   // Mqtt Topic
	Payload []byte `y3:"0x21" json:"payload"` // Mqtt Payload
	Time    int64  `y3:"0x22" json:"time"`    // Timestamp (ms)
	From    string `y3:"0x23" json:"from"`    // Source IP
}

func main() {
	webSocketServer, err = newWebSocketServer()
	if err != nil {
		log.Printf("❌ Initialize the socket.io server failure with err: %v", err)
		return
	}

	// sink server which will receive the data from `yomo-zipper`.
	go connectToZipper(zipperAddr)

	// serve socket.io server.
	go webSocketServer.Serve()
	defer webSocketServer.Close()

	router := gin.New()
	router.Use(ginMiddleware())
	router.GET("/socket.io/*any", gin.WrapH(webSocketServer))
	router.POST("/socket.io/*any", gin.WrapH(webSocketServer))
	router.StaticFS("/public", http.Dir("./sink/asset"))
	router.Run(WebSocketAddr)

	log.Print("✅ Serving socket.io on ", WebSocketAddr)
	err = http.ListenAndServe(WebSocketAddr, nil)
	if err != nil {
		log.Printf("❌ Serving the socket.io server on %s failure with err: %v", WebSocketAddr, err)
		return
	}
}

func newWebSocketServer() (*socketio.Server, error) {
	log.Print("Starting socket.io server...")
	server, err := socketio.NewServer(nil)
	if err != nil {
		return nil, err
	}

	// add all connected user to the room "yomo-demo".
	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		log.Print("connected:", s.ID())
		s.Join(WebSocketRoom)

		return nil
	})

	return server, nil
}

func ginMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestOrigin := c.Request.Header.Get("Origin")

		c.Writer.Header().Set("Access-Control-Allow-Origin", requestOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, Content-Length, X-CSRF-Token, Token, session, Origin, Host, Connection, Accept-Encoding, Accept-Language, X-Requested-With")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Request.Header.Del("Origin")

		c.Next()
	}
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

// connectToZipper connects to `yomo-zipper` and receives the real-time data.
func connectToZipper(zipperAddr string) error {
	urls := strings.Split(zipperAddr, ":")
	if len(urls) != 2 {
		return fmt.Errorf(`❌ The format of url "%s" is incorrect, it should be "host:port", f.e. localhost:9000`, zipperAddr)
	}
	host := urls[0]
	port, _ := strconv.Atoi(urls[1])
	cli, err := client.NewServerless(appName).Connect(host, port)
	if err != nil {
		return err
	}
	defer cli.Close()

	if enableDebug {
		cli.EnableDebug()
	}

	cli.Pipe(Handler)
	return nil
}

// Handler receives the real-time data and broadcast it to socket.io clients.
func Handler(rxstream rx.RxStream) rx.RxStream {
	stream := rxstream.
		Subscribe(SinkDataKey). // 监听来源于shake-flow生成的数据
		OnObserve(decode).      // 解码数据为结构体
		Map(broadcastData)      // 通过WebSocket广播数据

	return stream
}

func decode(v []byte) (interface{}, error) {
	// decode the data via Y3 Codec.
	data := SinkData{}
	err := y3.ToObject(v, &data)
	if err != nil {
		log.Print(err)
		return nil, err
	}

	return data, nil
}

// broadcastData broadcasts the real-time data to socket.io clients.
func broadcastData(_ context.Context, data interface{}) (interface{}, error) {
	broadcastS07 := func(sinkData SinkData) {
		s07Data := S07Data{Topic: sinkData.Topic, Time: sinkData.Time, From: sinkData.From}
		err := json.Unmarshal(sinkData.Payload, &s07Data)
		if err != nil {
			fmt.Printf("Unmarshal Payload error: %v\n", err)
		}
		fmt.Println(fmt.Sprintf("broadcast %s data: %v", WebSocketRoom, s07Data))
		webSocketServer.BroadcastToRoom("", WebSocketRoom, "receive_sink_s07", s07Data)
	}

	broadcastS05 := func(sinkData SinkData) {
		s05Data := S05Data{Topic: sinkData.Topic, Time: sinkData.Time, From: sinkData.From}
		err := json.Unmarshal(sinkData.Payload, &s05Data)
		if err != nil {
			fmt.Printf("Unmarshal Payload error: %v\n", err)
		}
		fmt.Println(fmt.Sprintf("broadcast %s data: %v", WebSocketRoom, s05Data))
		webSocketServer.BroadcastToRoom("", WebSocketRoom, "receive_sink_s05", s05Data)
	}

	if webSocketServer != nil && data != nil {
		sinkData := data.(SinkData)
		switch sinkData.Topic {
		case S07Topic:
			broadcastS07(sinkData)
		case S05Topic:
			broadcastS05(sinkData)
		}

		//fmt.Println(fmt.Sprintf("broadcast %s data: %v", WebSocketRoom, data))
		//webSocketServer.BroadcastToRoom("", WebSocketRoom, "receive_sink", data)
	} else {
		log.Printf("❌ Not eligible for broadcasting. webSocketServer=%v, data=%v\n", webSocketServer, data)
	}

	return data, nil
}

type S07Data struct {
	Topic string `json:"topic"` // Mqtt Topic
	Time  int64  `json:"time"`  // Timestamp (ms)
	From  string `json:"from"`  // Source IP

	TenantId    string  `json:"tenantId"`
	DevEui      string  `json:"devEui"`
	Data        string  `json:"data"`
	CreateDate  int64   `json:"createDate"`
	Temperature float64 `json:"temperature"`
	Vertical    float64 `json:"vertical"`
	Transverse  float64 `json:"transverse"`
}

type S05Data struct {
	Topic string `json:"topic"` // Mqtt Topic
	Time  int64  `json:"time"`  // Timestamp (ms)
	From  string `json:"from"`  // Source IP

	TenantId   string `json:"tenantId"`
	DevEui     string `json:"devEui"`
	Data       string `json:"data"`
	CreateDate int64  `json:"createDate"`
	Key        string `json:"key"`
}
