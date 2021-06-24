package main

import (
	"context"
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
)

var (
	webSocketServer *socketio.Server
	err             error
	appName         = getEnvString("SHAKE_SINK_APP_NAME", "shake-sink")
	zipperAddr      = getEnvString("SHAKE_ZIPPER_ADDR", "localhost:9000")
)

type SinkData struct {
	Topic   string `y3:"0x20"` // Mqtt Topic
	Payload []byte `y3:"0x21"` // Mqtt Payload
	Time    int64  `y3:"0x22"` // Timestamp (ms)
	From    string `y3:"0x23"` // Source IP
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
	router.StaticFS("/public", http.Dir("./asset"))
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

	cli.Pipe(Handler)
	return nil
}

// Handler receives the real-time data and broadcast it to socket.io clients.
func Handler(rxstream rx.RxStream) rx.RxStream {
	stream := rxstream.
		Subscribe(SinkDataKey). // 监听来源于shake-flow生成的数据
		OnObserve(decode).      // 解码数据为结构体
		Map(broadcastData).     // 通过WebSocket广播数据
		Encode(SinkDataKey)

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
	if webSocketServer != nil && data != nil {
		fmt.Println(fmt.Sprintf("broadcast %s data: %v", WebSocketRoom, data))
		webSocketServer.BroadcastToRoom("", WebSocketRoom, "receive_sink", data)
	}

	return data, nil
}
