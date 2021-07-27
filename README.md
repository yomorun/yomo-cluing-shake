# YoMo在CluingOS中的实践

## 前言

[YoMo](https://github.com/yomorun/yomo) 是一个开源编程框架，为边缘计算领域的低时延流式数据处理而打造，它底层基于 HTTP 3.0 的核心通讯层 IETF QUIC 协议通讯，以 Functional Reactive Programming 为编程范式，方便开发者构建可靠、安全的时序型数据的实时计算应用，并针对5G和WiFi-6场景优化，释放实时计算价值。

[Y3](https://github.com/yomorun/y3-codec-golang)是一种[YoMo Codec](https://github.com/yomorun/yomo-codec)的Golang实现，它描述了一个快速和低CPU损耗的编解码器，专注于边缘计算和流处理。查看 [explainer](https://github.com/yomorun/y3-codec-golang/blob/master/explainer_CN.md) 获取更多信息，了解更多与[YoMo](https://github.com/yomorun/yomo)组合的方式。

[CluingOS](https://www.cluing.com.cn/) 是一款以 Kubernetes 为内核的云原生超融合工业物联平台，它的架构可以非常方便地使第三方应用与云原生生态组件进行集成、整合和安装，支持云原生应用在多云与多集群的统一分发和运维管理。。

在这个案例里，我们结合了**YoMo+Y3**的低延时流式处理与**CluingOS**分布式部署的特性，展现出如何开发部署一套高效的工业数据收集应用系统，体验从边缘端收集传感器数据，低延时高效地跨越2000多公里地传输到云端进行数据流式处理的全过程，基于这个案例你可以照葫芦画瓢地开发出满足自已需求的应用场景。

## 架构

![YoMo_CluingOS](https://github.com/yomorun/yomo-cluing-shake/releases/download/v0.1.0/YoMo_CluingOS.png)

> 关于本案例更详细的信息，请阅读此文章: [YoMo在CluingOS中的实践](https://github.com/yomorun/yomo-cluing-shake/blob/main/BLOG.md) 或者 [Wiki](https://github.com/yomorun/yomo-cluing-shake/wiki/YoMo%E5%9C%A8CluingOS%E4%B8%AD%E7%9A%84%E5%AE%9E%E8%B7%B5)。



## 如何本地运行

### 先决条件

[Install Go](https://golang.org/doc/install)

注意：如果发现有下载安装不了的情况，需要配置一下自已http代理:

```bash
export http_proxy=http://{ProxyIP}:{ProxyPort};export https_proxy=http://{ProxyIP}:{ProxyPort};
```

### 1. Clone Repository

```bash 
$ git clone https://github.com/yomorun/yomo-cluing-shake.git
$ go get ./...
```

### 2. 安装YoMo CLI

```bash
$ go install github.com/yomorun/cli/yomo@latest
```

执行下面的命令，确保yomo已经在环境变量中，有任何问题请参考 [YoMo 的详细文档](https://github.com/yomorun/yomo)

```bash
$ yomo version
YoMo CLI version: v0.0.7
```

也可以直接下载可执行文件: [yomo-v0.0.7-x86_64-linux.tgz](https://github.com/yomorun/cli/releases/download/v0.0.7/yomo-v0.0.7-x86_64-linux.tgz)， [yomo-v0.0.7-aarch64-linux.tgz](https://github.com/yomorun/cli/releases/download/v0.0.7/yomo-v0.0.7-aarch64-linux.tgz)

### 3. 运行shake-zipper

```bash
$ yomo serve -c ./zipper/workflow.yaml
# 如果是本地调试，可以运行 `make debug_zipper`
```

### 4. 运行shake-flow

```bash
$ go run ./flow/app.go
# 如果是本地调试，可以运行 `make debug_flow`
```

### 5. 运行shake-sink

```bash
$ go run ./sink/main.go
# 如果是本地调试，可以运行 `make debug_sink`
# 访问 http://localhost:8000/public/ 查看效果
```

### 6. 运行shake-source

```bash
$ CONNECTOR_MQTT_AUTH_ENABLE=true \
CONNECTOR_MQTT_AUTH_USERNAME=yomo \
CONNECTOR_MQTT_AUTH_PASSWORD=yomo \
go run ./source/main.go
# 如果是本地调试，可以运行 `make debug_source`
```

### 7. 运行emitter

```bash
$ SHAKE_SOURCE_MQTT_PUB_INTERVAL=2000 go run ./cmd/emitter/main.go
# 如果是本地调试，可以运行 `make debug_emitter`
```

### 8. 运行shake-web

```bash
$ cd web
$ yarn
$ yarn start
# 如果是本地调试，可以运行 `make debug_web`
```

访问 http://localhost:3000/ 查看效果。

## 如何打包镜像

以打包`yomorun/shake-source:latest`为例：

```bash
# 构建本地镜像
$ docker build --no-cache -f Dockerfile.source -t local/shake-source:latest .
# 标记为你要发布的镜像
$ docker tag local/shake-source:latest yomorun/shake-source:latest
# 发布镜像
$ docker login -u yomorun -p {你的密码}
$ docker push yomorun/shake-source:latest
$ docker logout
```

源码已经为你提供了打包镜像所需的Dockerfile文件，打包出如下镜像：

- yomorun/shake-zipper:latest
- yomorun/shake-flow:latest
- yomorun/shake-sink:latest
- yomorun/shake-source:latest
- yomorun/shake-emitter:latest
- yomorun/shake-web:latest
- yomorun/shake-web-proxy:latest



## 如何通过CluingOS发布服务

请参考相关文档:[CluingOS部署](https://github.com/yomorun/yomo-cluing-shake/blob/main/BLOG.md#cluingos%E9%83%A8%E7%BD%B2)，主要涉及环境变量和端口的配置，列出可配置的变量如下：

| 服务              | 环境变量                                                     | 暴露端口  |
| ----------------- | ------------------------------------------------------------ | --------- |
| shake-zipper      |                                                              | 9000/UDP  |
| shake-flow        | SHAKE_ZIPPER_ADDR={shake-zipper的地址}                       |           |
| shake-sink        | SHAKE_ZIPPER_ADDR={shake-zipper的地址}                       | 8000/TCP  |
| shake-source      | SHAKE_ZIPPER_ADDR={shake-zipper的地址}<br />SHAKE_SOURCE_SERVER_ADDR=0.0.0.0:1883<br />CONNECTOR_MQTT_AUTH_ENABLE=true<br />CONNECTOR_MQTT_AUTH_USERNAME=yomo<br />CONNECTOR_MQTT_AUTH_PASSWORD=yomo | 1883/TCP  |
| shake-web         | HOST=0.0.0.0<br />REACT_APP_WEB_SOCKET_URL=http://{shake-sink ip}:{shake-sink port} | 3000/HTTP |
| *shake-emitter*   | *SHAKE_SOURCE_MQTT_BROKER_ADDR=tcp://{shake-source ip}:{shake-source port}* |           |
| *shake-web-proxy* | *PROXY_PASS=http://{shake-web cpe-vpn-ip}:{shake-web port}*  | 8989/HTTP |



## 运行效果

![effect](https://github.com/yomorun/yomo-cluing-shake/releases/download/v0.1.0/effect.gif)
