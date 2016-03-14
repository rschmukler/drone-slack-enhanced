.PHONY: docker

EXECUTABLE ?= drone-slack
IMAGE ?= rschmukler/drone-slack

docker:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(EXECUTABLE)
	docker build --rm -t $(IMAGE) .
