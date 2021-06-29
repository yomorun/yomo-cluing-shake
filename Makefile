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
	go run ./source/main.go

debug_emitter:
	go run ./cmd/emitter/main.go

debug_web:
	cd web && yarn && yarn start

