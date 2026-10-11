
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server
	docker build . -t http_example:latest
	rm -rf server