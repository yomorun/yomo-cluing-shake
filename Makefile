VERSION=`git describe --tags`
BUILD=`date -u +.%Y%m%d.%H%M%S`
VER ?= $(shell cat VERSION)

debug_zipper:
	yomo serve -c ./zipper/workflow.yaml

debug_flow:
	go run ./flow/app.go

debug_sink:
	go run ./sink/main.go #http://localhost:8000/public/

debug_source:
	CONNECTOR_MQTT_AUTH_ENABLE=true CONNECTOR_MQTT_AUTH_USERNAME=yomo CONNECTOR_MQTT_AUTH_PASSWORD=yomo go run ./source/main.go

debug_emitter:
	SHAKE_SOURCE_MQTT_PUB_INTERVAL=2000 go run ./cmd/emitter/main.go

debug_web:
	cd web && yarn && HOST=0.0.0.0 REACT_APP_WEB_SOCKET_URL=http://localhost:8000 yarn start

debug_homey_source:
	SHAKE_ZIPPER_ADDR=yomo.cluing.com:32703 go run ./source/main.go
	#SHAKE_ZIPPER_ADDR=10.0.100.3:32703 go run ./source/main.go

debug_homey_web:
	cd web && yarn && HOST=0.0.0.0 REACT_APP_WEB_SOCKET_URL=http://yomo.cluing.com:30095 yarn start

debug_homey_emitter:
	SHAKE_SOURCE_MQTT_PUB_INTERVAL=2000 go run ./cmd/emitter/main.go

