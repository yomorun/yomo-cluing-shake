# yomo-cluing-shake

![effect](https://github.com/yomorun/yomo-cluing-shake/releases/download/v0.1.0/effect.gif)



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
$ go run ./source/main.go
# 如果是本地调试，可以运行 `make debug_source`
```

### 7. 运行emitter

```bash
$ go run ./cmd/emitter/main.go
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

请参考相关文档:[在CluingOS中部署yomo-cluing-shake](https://ycella.feishu.cn/file/boxcnaMigXT8nvLJuh7ZnQm2Xkg)，主要涉及环境变量和端口的配置，列出可配置的变量如下：

| 服务            | 环境变量                                                     | 暴露端口  |
| --------------- | ------------------------------------------------------------ | --------- |
| shake-zipper    |                                                              | 9000/UDP  |
| shake-flow      | SHAKE_ZIPPER_ADDR={shake-zipper的地址}                       |           |
| shake-sink      | SHAKE_ZIPPER_ADDR={shake-zipper的地址}                       | 8000/TCP  |
| shake-source    | SHAKE_ZIPPER_ADDR={shake-zipper的地址}<br />SHAKE_SOURCE_SERVER_ADDR=0.0.0.0:1883<br />CONNECTOR_MQTT_AUTH_ENABLE=true<br />CONNECTOR_MQTT_AUTH_USERNAME=yomo<br />CONNECTOR_MQTT_AUTH_PASSWORD=yomo | 1883/TCP  |
| shake-web       | HOST=0.0.0.0<br />REACT_APP_WEB_SOCKET_URL=http://{shake-sink ip}:{shake-sink port} | 3000/HTTP |
| shake-emitter   | SHAKE_SOURCE_MQTT_BROKER_ADDR=tcp://{shake-source ip}:{shake-source port} |           |
| shake-web-proxy | PROXY_PASS=http://{shake-web cpe-vpn-ip}:{shake-web port}    | 8989/HTTP |

