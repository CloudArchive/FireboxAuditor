APP_NAME := firebox-auditor

.PHONY: all frontend build-linux build-windows build-darwin build-all docker clean

all: frontend build-darwin

frontend:
	cd frontend && npm install && npm run build

build-linux: frontend
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/$(APP_NAME)-linux-amd64 .

build-windows: frontend
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/$(APP_NAME)-windows-amd64.exe .

build-darwin: frontend
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/$(APP_NAME)-darwin-arm64 .

build-all: build-linux build-windows build-darwin

docker:
	docker build -t $(APP_NAME) .

clean:
	rm -rf dist/ static/

dev:
	cd frontend && npm run dev &
	go run .
