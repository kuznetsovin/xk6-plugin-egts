all: build run

build:
	go test ./... && xk6 build --output bin/k6 --with xk6-plugin-dtm="$(shell pwd)"

run:
	bin/k6 run example/example.js
