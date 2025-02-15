

all: build/xip8-cli build/xip8-gui build/xip8-web

build/xip8-cli: *.go go.sum
	go build -o build/xip8-cli ./cmd/cli/*

build/xip8-gui: *.go go.sum
	go build -o build/xip8-gui ./cmd/gui/*

build/xip8-web: *.go go.sum
	go build -o build/xip8-web ./cmd/web/*