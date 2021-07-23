# YoMo在CluingOS中的实践

## 前言

[YoMo](https://github.com/yomorun/yomo) 是一个开源编程框架，为边缘计算领域的低时延流式数据处理而打造，它底层基于 HTTP 3.0 的核心通讯层 IETF QUIC 协议通讯，以 Functional Reactive Programming 为编程范式，方便开发者构建可靠、安全的时序型数据的实时计算应用，并针对5G和WiFi-6场景优化，释放实时计算价值。

[Y3](https://github.com/yomorun/y3-codec-golang)是一种[YoMo Codec](https://github.com/yomorun/yomo-codec)的Golang实现，它描述了一个快速和低CPU损耗的编解码器，专注于边缘计算和流处理。查看 [explainer](https://github.com/yomorun/y3-codec-golang/blob/master/explainer_CN.md) 获取更多信息，了解更多与[YoMo](https://github.com/yomorun/yomo)组合的方式。

[CluingOS](https://www.cluing.com.cn/) 是一款以 Kubernetes 为内核的云原生超融合工业物联平台，它的架构可以非常方便地使第三方应用与云原生生态组件进行集成、整合和安装，支持云原生应用在多云与多集群的统一分发和运维管理。。

在这个案例里，我们结合了**YoMo+Y3**的低延时流式处理与**CluingOS**分布式部署的特性，展现出如何开发部署一套高效的工业数据收集应用系统，体验从边缘端收集传感器数据，低延时高效地跨越2000多公里地传输到云端进行数据流式处理的全过程，基于这个案例你可以照葫芦画瓢地开发出满足自已需求的应用场景。

## 述语

- xxx-source：表示一个数据源收集程序，能直接接收MQTT协议的数据。
- xxx-zipper：表示一个工作流和控制平面。
- xxx-flow：表示一个工作流单元，用于实际的业务逻辑处理，被zipper调度。
- xxx-sink：表示一个数据的传送目的地，本案例为一个消费数据的WebSocket服务，被zipper调度。
- xxx-web：表示一个展示实时传感数据的Web服务。

## 架构

![YoMo_CluingOS](https://github.com/yomorun/yomo-cluing-shake/releases/download/v0.1.0/YoMo_CluingOS.png)

从图中的案例可见，区分了边缘端和云端两个独立的服务区域，其中边缘端位于上海，云端位于广州，距离相隔2000多公里，我们可以在最后的测试中看到其低延时的流式处理为数据的收集和处理提供了令人惊喜的优化。另外，CluingOS超融合工业物联平台的容器化分布式部署能高效地部署调试我们的应用，真香！下面简单介绍一下各个模块和服务：

### 传感器

> 震动传感器，薄膜按键传感器

- 震动传感器。用于产生震动相关的原始数据，经过Lora接收器转为MQTT协议的数据，数据格式如下：

  - TOPIC：shake/20210627_cluing/S07

  - Payload：

    ```json
    {
    	// 租户数据库实例
    	"tenantId": "20210627_cluing",
    	// 采集设备终端DEVEUI
    	"devEui": "0850533277387820",
    	// 采集原始数据
    	"data": "CwcMys+69Iks0As4YS4N6A==",
    	// 采集数据时间
    	"createDate": 1624937248919,
    	// 采集温度
    	"temperature": 75,
    	// Z轴振动强度
    	"vertical": 81,
    	//X轴振动强度
    	"transverse": 53
    }
    ```

- 薄膜按键传感器。用于产生按键相关的原始数据，经过Loar接收器转为MQTT协议的数据，数据格式如下：

  - TOPIC：shake/20210627_cluing/S05

  - Payload:

    ```json
    {
      // 租户数据库实例
    	"tenantId": "20210627_cluing",
      // 采集设备终端DEVEUI
    	"devEui": "393235307d377504",
      // 采集原始数据
    	"data": "AAAQ5gAAAQIIAA==",
      // 采集数据时间
    	"createDate": 1624937248919,
      // 按键设备按键值
    	"key": "0800"
    }
    ```

    

### 5G CPE 一体机

> Lora接收器，转发引擎，shake-source，shake-web

`5G CPE 一体机`是部署在边缘端的一个网关设备，它可以接收传感器的数据并转换为MQTT协议的数据，我们的YoMo边缘端的接收器shake-source就是部署在这个网关设备上，它的一大特色是可以接受CluingOS超融合工业物联平台的容器部署和资源调度，这样即使你在相隔千里的地方也很容易的分发应用到这个网关设备上，并不需要远程登录进行操作哦。

- Lora接收器是网关的默认服务。可以接收各种各样的传感器的监控数据，并转换为MQTT协议，不同设备的数据可以分配置不同的主题易于管理，上节中的震动传感器和薄膜按键传感器的数据就被转换为定义所示的数据格式了。
- 转发引擎是网关的默认服务。可以把MQTT的数据转发给不同的MQTT Broker服务，当然也包括我们案例中的shake-source。
- shake-source数据源。基于[YoMo](https://github.com/yomorun/yomo)框架开发的数据源接收服务，它的作用是把MQTT协议的数据转换为[Y3](https://github.com/yomorun/y3-codec-golang)数据格式并以[QUIC](https://datatracker.ietf.org/wg/quic/documents/)的方式传输到云端的shake-zipper工作流引擎。
- shake-web数据展示。这是一个展示两种传感器的实时数据的展示Web服务，主要是消费shake-sink提供的WebSocket数据，同时也可以展示一个完整的RTT往返时延的实时性。

### SaaS服务

> shake-zipper，shake-flow，shake-sink，HOMEY呼美事件精益化管理系统

这是一套完整的云端服务，其中的容器化部署也是受到了CluingOS的调度和管理，只需要以不同的用户登录CluingOS就可以切换管理在边缘端或者在云端的服务。

- shake-zipper工作流引擎。通过编排(workflow.yaml)可以调度多个flow和sink，让它们以流的方式把业务逻辑串联起来，以满足复杂的需求。与之相连的所有通信和编解码均以QUIC+Y3进行，提供可靠实时的流式处理，全程体验流式编程的乐趣。
- shake-flow逻辑处理单元。在这个案体中，处理单元把从source传输过来的数据解码为Topic和Playload后分别处理震动传感器和薄膜按键传感器两种设备的数据，并且达到一定的阀值后调用HOMEY呼美管理系统进行警报或者下发控制边缘端设备的控制指令。
- shake-sink数据输出单元。在这个案例中并没有输出到数据库，而是通过搭建一个WebSocket服务器，把实时的传感器数据输出给任意的网页进行展示消费。这里的数据是作为shake-web的数据源进行展示。
- HOMEY呼美管理系统。呼美系统接收到shake-flow的事件通知后会发出警报信息或者下发控制指令到边缘端控制某些设备实施某种操作。但在这个案例中，我们会把在呼美系统接收到事件通知的时间点用来作为延时的统计终点，分析我们基于YoMo+Y3的低延时确实得到很大的优化。

### CluingOS超融合工业物联平台

> 以不同用户登录可以切换对边缘端服务或者云端服务的部署和管理

CluingOS提供工业物联网大数据智能平台服务及容器化、订阅式、微服务架构的“现场协同＋流程管控＋数据智能”的端到端一体化透明工厂系统，支持私有云、公有云或混合云的多种方式、分布式部署实施。

## 代码

项目[yomo-cluing-shake](https://github.com/yomorun/yomo-cluing-shake)提供了全套的源代码，下表提供了每个模块的简要说明，供感兴趣的朋友查看，参照这个案体的代码，可以轻松开发出类拟场景的案例。

| 模块      | 地址                                                         | 本地运行             | 说明                            |
| --------- | ------------------------------------------------------------ | -------------------- | ------------------------------- |
| zipper    | [zipper](https://github.com/yomorun/yomo-cluing-shake/tree/main/zipper) | `make debug_zipper`  | 编排本案例的工作流和数据流向    |
| flow      | [flow](https://github.com/yomorun/yomo-cluing-shake/tree/main/flow) | `make debug_flow`    | 对传感器数据进行预处理和警报    |
| sink      | [sink](https://github.com/yomorun/yomo-cluing-shake/tree/main/sink) | `make debug_sink`    | 提供WebSocket服务用于数据展示   |
| source    | [source](https://github.com/yomorun/yomo-cluing-shake/tree/main/source) | `make debug_source`  | 收集MQTT消息格式的传感器数据    |
| emitter   | [emitter](https://github.com/yomorun/yomo-cluing-shake/tree/main/cmd/emitter) | `make debug_emitter` | 模拟产生震动和按键数据          |
| web       | [web](https://github.com/yomorun/yomo-cluing-shake/tree/main/web) | `make debug_web`     | 消费WebSocket服务展示传感器数据 |
| quic-mqtt | [yomo-source-mqtt-starter](https://github.com/yomorun/yomo-source-mqtt-starter) |                      | 开发xxx-source的通用组件        |

## 容器化部署

通过下载上节的项目代码可以快速地本地运行，体验YoMo开发的乐趣，同时我们提供了各个模块对应的Dockerfile文件用于打包对应的镜像，并且上传到hub.dockder.com，供CluingOS进行部署。

| 模块      | Dockerfile                                                   | 镜像地址                                                     | 最新版本                     |
| --------- | ------------------------------------------------------------ | ------------------------------------------------------------ | ---------------------------- |
| zipper    | [Dockerfile.zipper](https://github.com/yomorun/yomo-cluing-shake/blob/main/Dockerfile.zipper) | [shake-zipper](https://hub.docker.com/r/yomorun/shake-zipper) | yomorun/shake-zipper:latest  |
| flow      | [Dockerfile.flow](https://github.com/yomorun/yomo-cluing-shake/blob/main/Dockerfile.flow) | [shake-flow](https://hub.docker.com/r/yomorun/shake-flow)    | yomorun/shake-flow:latest    |
| sink      | [Dockerfile.sink](https://github.com/yomorun/yomo-cluing-shake/blob/main/Dockerfile.sink) | [shake-sink](https://hub.docker.com/r/yomorun/shake-sink)    | yomorun/shake-sink:latest    |
| source    | [Dockerfile.source](https://github.com/yomorun/yomo-cluing-shake/blob/main/Dockerfile.source) | [noise-source](https://hub.docker.com/r/yomorun/noise-source) | yomorun/shake-source:latest  |
| emitter   | [Dockerfile.emitter](https://github.com/yomorun/yomo-cluing-shake/blob/main/Dockerfile.emitter) | [shake-emitter](https://hub.docker.com/r/yomorun/shake-emitter) | yomorun/shake-emitter:latest |
| web       | [Dockerfile.web](https://github.com/yomorun/yomo-cluing-shake/blob/main/Dockerfile.web) | [shake-web](https://hub.docker.com/r/yomorun/shake-web)      | yomorun/shake-web:latest     |
| quic-mqtt |                                                              | [yomorun/quic-mqtt](https://www.oschina.net/action/GoToLink?url=https%3A%2F%2Fhub.docker.com%2Fr%2Fyomorun%2Fquic-mqtt) | yomorun/quic-mqtt:latest     |

yomorun/quic-mqtt:latest 是打包xxx-source的基础镜像，可以快速打包自定义代码，但本案例中可以暂时忽略。

## CluingOS部署

> 通过不同的用户登录`CluingOS工业超融合系统`可以分别管理**边缘端**和**云端**的容器。

### 部署云端服务

> 用户A部署zipper/flow/sink服务到云机器。

1. #### 创建自定义应用。

   > 创建一个自制应用shake-cloud，用于管理云端的服务。

![cluing_cloud_1](https://github.com/yomorun/yomo-cluing-shake/releases/download/v0.1.0/cluing_cloud_1.jpg)



2. #### 创建服务。

   > 进入shake-cloud应用的控制台，选择添加服务，选择无状态服务，则进入创建服务的流程。

   ![cluing_cloud_2](https://github.com/yomorun/yomo-cluing-shake/releases/download/v0.1.0/cluing_cloud_2.jpg)

   - 基本信息：指定服务的名称为shake-sink

   - 容器镜像：
     - 添加容器镜像：选择从DockerHub中搜索shake-sink
       ![cluing_cloud_3](https://github.com/yomorun/yomo-cluing-shake/releases/download/v0.1.0/cluing_cloud_3.jpg)
     - 端口设置：指定容器暴露的服务端口，例如：8000
     - 环境变量：例如`SHAKE_ZIPPER_ADDR=shake-zipper.yomo-cluing-shake:9000`，这里的`shake-zipper.yomo-cluing-shake`则时创建shake-zipper后获得的zipper在内部DNS名。
   - 挂载存储和高级设置：在这个案例中都不需要设置。
   - 编辑外网访问：选择NodePort的访问方式，获取得到对外暴露端口号30095
     ![cluing_cloud_4](https://github.com/yomorun/yomo-cluing-shake/releases/download/v0.1.0/cluing_cloud_4.jpg)

3. #### 服务列表

   > 至此，分别创建了zipper/flow/sink的服务。

   ![cluing_cloud_5](https://github.com/yomorun/yomo-cluing-shake/releases/download/v0.1.0/cluing_cloud_5.jpg)

   

### 部署边缘端服务

> 用户B部署source/web服务到边缘的5GCPE一体机。

部署方式与`部署云端服务`相同，先创建一个shake-edge应用，然后在应用中创建对应的无状态服务，获得服务列表：
![cluing_cloud_6](https://github.com/yomorun/yomo-cluing-shake/releases/download/v0.1.0/cluing_cloud_6.jpg)

### 接入传感器展示效果

> 转发引擎把MQTT的消息转发到shake-souce服务。

为了接入传感器数据，只需要修改**转发引擎**的转发地址为**shake-source**服务的地址端口，即可通过**shake-web**展示收到的实时数据。

![cluing_cloud_7](https://github.com/yomorun/yomo-cluing-shake/releases/download/v0.1.0/cluing_cloud_7.jpg)

## 效果对比

> 为了与传统http上报数据进行时延对比，设计测试用例进行效果对比。

本次与yomo集成均采用真实环境进行，所有相关应用组件全部采用CluingOS进行容器化部署和安装，场景覆盖了云-边-端的应用。除了验证yomo的加速效果，还将yomo集成整合进了凌犀平台体系（CluingOS/AIOT/MOM），以下主要介绍yomo的加速测试效果。

### 实际场景

![cluing_cloud_8](https://github.com/yomorun/yomo-cluing-shake/releases/download/v0.1.0/cluing_cloud_8.png)

从测试架构图可见，使用两组测试进行对比：

#### QUIC+Y3通道

如图中橙色流程所示，数据传输路径为：

![cluing_cloud_9](https://github.com/yomorun/yomo-cluing-shake/releases/download/v0.1.0/cluing_cloud_9.png)

#### HTTP通道

如图中绿色流程所示，数据传输路径为：

![cluing_cloud_10](https://github.com/yomorun/yomo-cluing-shake/releases/download/v0.1.0/cluing_cloud_10.png)

### 环境说明

- 5G CPE一体机放置在上海金桥办公室。
-  MOM智造运营管理系统部署在广州腾讯云。
- 测试时延链路为**上海金桥办公室**至**广州腾讯云**这段网络。

### 测试方法

- 准备传感器A（3932353062376611）和传感器B（0850533277387820）。
- 传感器A走传统**HTTP**协议、传感器B走**QUIC+Y3**协议。
- 每个传感器传输600条记录，求出平均时延。

### 计算公式

```sql
select device_sn,(timeCount/count)*1000 agv,count from(
select device_sn,sum(unix_timestamp(end_time) -unix_timestamp(begin_time)) as timeCount,count(id) as count from sdm_device_logs_copy2 GROUP BY device_sn
) t
```

### 测试结论

| 序号 | 协议类型 | 设备编码                    | 平均时延（毫秒） | 测试记录数 |
| ---- | -------- | --------------------------- | ---------------- | ---------- |
| 1    | HTTP     | 传感器A（3932353062376611） | 100.268333(ms)   | 600        |
| 2    | QUIC+Y3  | 传感器B（0850533277387820） | 33.493333(ms)    | 600        |

从采集样本中利用计算公式，求得每个传感器传输600条数据的平均时延，可以看出**HTTP协议**平均时延为**100.26毫秒**，**QUIC+Y3协议**平均时延只有**33.49毫秒**，具有非常明显的加速效果。

## 结束语

近年来，新一代信息技术发展突飞猛进，互联网由消费领域向工业领域加速拓展。从数字产业化方面来看，工业互联网想要向更大范围、更深程度和更高水平发展，亟需新的技术、产品和解决方案。[YoMo](https://github.com/yomorun/yomo) 开源编程框架能够大幅提升从边端到云端的传输效率，提升实时性和获得低延时的优势，同时全新的流式计算和编程范式给开发者一个全新的开发体验，更自然更高效地开发出流式计算的应用。借助[凌犀](https://www.cluing.com.cn/)的CluingOS超融合工业物联平台可以快速方便地把容器服务部署在边缘端和云端，实现服务治理能力的大幅提升。
