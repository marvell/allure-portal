BUILD_ENVPARMS:=CGO_ENABLED=0 CC=gcc

.PHONY: all
all: build

.PHONY: build
build:
	$(BUILD_ENVPARMS) go build -o ./bin/allure-portal

.PHONY: clean
clean:
	rm -rf storage/test

.PHONY: fmt
fmt:
	gofmt -s -w *.go

.PHONY: test
test:
	go test -cover ./...

.PHONY: docker-build
docker-build:
	docker build -t marvell/allure-portal .

.PHONY: docker-push
docker-push:
	docker push marvell/allure-portal
