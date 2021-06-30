package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/yomorun/y3-codec-golang"

	"github.com/yomorun/yomo/pkg/rx"

	"github.com/yomorun/yomo/pkg/client"
)

const (
	ShakeDataKey = 0x10
	SinkDataKey  = 0x11
	S07Topic     = "shake/20210627_cluing/S07"
	S05Topic     = "shake/20210627_cluing/S05"
)

var (
	appName               = getEnvString("SHAKE_FLOW_APP_NAME", "shake-flow")
	zipperAddr            = getEnvString("SHAKE_ZIPPER_ADDR", "localhost:9000")
	enableDebug           = getEnvBool("SHAKE_FLOW_ENABLE_DEBUG", false)
	enableDispatch        = getEnvBool("SHAKE_FLOW_ENABLE_DISPATCH", true)
	receiveGatewayInfoUrl = getEnvString("RECEIVE_GATEWAY_INFO", "http://yomo.cluing.com:30558/api-sdm/v1/receiveGatewayInfo")

	homeyService = newHomeyService()
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
	if enableDebug {
		cli.EnableDebug()
	}
	cli.Pipe(Handler)
}

func Handler(rxStream rx.RxStream) rx.RxStream {
	stream := rxStream.
		Subscribe(ShakeDataKey). // 监听来自shake-source的数据
		OnObserve(decode).       //解码数据到结构体
		Map(dispatch).           //把数据发送到呼美系统
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
	fmt.Println(fmt.Sprintf("[%s] %d > topic: %s; payload.len:%d ⚡️=%dms", mold.From, mold.Time, mold.Topic, len(mold.Payload), rightNow-mold.Time))

	return mold, nil
}

var dispatch = func(_ context.Context, i interface{}) (interface{}, error) {
	if !enableDispatch {
		return i, nil
	}

	value := i.(ShakeData)

	switch value.Topic {
	case S07Topic:
		go func() {
			err := handleS07(&value)
			if err != nil {
				fmt.Printf("❌ handleS07 error:%v\n", err)
			}
		}()
	case S05Topic:
		go func() {
			err := handleS05(&value)
			if err != nil {
				fmt.Printf("❌ handleS05 error:%v\n", err)
			}
		}()
	}

	return value, nil
}

func handleS07(v *ShakeData) error {
	var mold S07
	err := json.Unmarshal(v.Payload, &mold)
	if err != nil {
		return err
	}

	request := GatewayRequest{
		Tenant:  mold.TenantId,
		SubDate: time.Now().Unix(),
		DevEUI:  mold.DevEui,
		Content: mold.Data,
	}

	response, err := homeyService.ReceiveGatewayInfo(&request)
	if err != nil || response == nil {
		return errors.New(fmt.Sprintf("ReceiveGatewayInfo error: Tenant=%s, SubDate=%d, DevEUI=%s, err=%v",
			request.Tenant, request.SubDate, request.DevEUI, err))
	}
	if response != nil && response.RespCode != 0 {
		return errors.New(fmt.Sprintf("error reponse: RespCode=%v, RespMsg=%v", response.RespCode, response.RespMsg))
	}

	fmt.Printf("✅ send S07 to HOMEY with: Temperature=%v, Vertical=%v, Transverse=%v\n",
		mold.Temperature, mold.Vertical, mold.Transverse)
	return nil
}

func handleS05(v *ShakeData) error {
	var mold S05
	err := json.Unmarshal(v.Payload, &mold)
	if err != nil {
		return err
	}

	request := GatewayRequest{
		Tenant:  mold.TenantId,
		SubDate: time.Now().Unix(),
		DevEUI:  mold.DevEui,
		Content: mold.Data,
	}

	response, err := homeyService.ReceiveGatewayInfo(&request)
	if err != nil || response == nil {
		return errors.New(fmt.Sprintf("ReceiveGatewayInfo error: Tenant=%s, SubDate=%d, DevEUI=%s, err=%v",
			request.Tenant, request.SubDate, request.DevEUI, err))
	}
	if response.RespCode != 0 {
		return errors.New(fmt.Sprintf("error reponse: RespCode=%v, RespMsg=%v", response.RespCode, response.RespMsg))
	}

	fmt.Printf("✅ send S05 to HOMEY with: Key=%v\n", mold.Key)
	return nil
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

type S07 struct {
	TenantId    string `json:"tenantId"`
	DevEui      string `json:"devEui"`
	Data        string `json:"data"`
	CreateDate  string `json:"createDate"`
	Temperature string `json:"temperature"`
	Vertical    string `json:"vertical"`
	Transverse  string `json:"transverse"`
}

type S05 struct {
	TenantId   string `json:"tenantId"`
	DevEui     string `json:"devEui"`
	Data       string `json:"data"`
	CreateDate string `json:"createDate"`
	Key        string `json:"key"`
}

type HomeyService interface {
	ReceiveGatewayInfo(request *GatewayRequest) (*GatewayResponse, error)
}

func newHomeyService() HomeyService {
	return &homeyServiceImpl{receiveGatewayInfoUrl: receiveGatewayInfoUrl}
}

type GatewayRequest struct {
	Tenant  string `json:"tenant"`  // 租户编码，MQTTJSON格式里的tenantId值, 必须。
	SubDate int64  `json:"subDate"` // 接收时间，必须。
	DevEUI  string `json:"devEUI"`  // 终端编码,MQTT JSON格式里的devEui值, 必须。
	Content string `json:"content"` // 终端采集与输入终端LORA原始值即MQTTJSON格式里的data值，不是必须。
}

type GatewayResponse struct {
	RespCode int         `json:"resp_code"` // 返回成功0：成功，必须。
	RespMsg  string      `json:"resp_msg"`  // 成功消息，必须。
	Datas    interface{} `json:"datas"`     // 不是必须。
}

type homeyServiceImpl struct {
	receiveGatewayInfoUrl string
}

func (h homeyServiceImpl) ReceiveGatewayInfo(request *GatewayRequest) (*GatewayResponse, error) {
	// 请求内容进行JSON序列化
	buf, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("#1 buf=%v\n", string(buf))

	// 构造请求参数
	req, err := http.NewRequest("POST", h.receiveGatewayInfoUrl, bytes.NewBuffer(buf))
	if err != nil {
		fmt.Printf("#1 NewRequest err=%v\n", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// 执行请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Do request error: %v", err))
	}
	defer resp.Body.Close()

	// 处理响应码
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("request fail, StatusCode: %v", resp.StatusCode))
	}

	// 读取响应内容
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("ReadAll Body error: %v", err))
	}

	// 解码为响应结果
	var result GatewayResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unmarshal GatewayResponse error: %v", err))
	}

	return &result, nil
}
