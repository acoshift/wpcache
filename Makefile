IMAGE=acoshift/wpcache
TAG=1.0.0
GOLANG_VERSION=1.8
REPO=github.com/acoshift/wpcache

wpcache: main.go
	go get -v
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-w -s' -o wpcache ./main.go

build:
	docker pull golang:$(GOLANG_VERSION)
	docker run --rm -it -v $(PWD):/go/src/$(REPO) -w /go/src/$(REPO) golang:$(GOLANG_VERSION) /bin/bash -c "make wpcache"
	docker build --pull -t $(IMAGE):$(TAG) .

push: clean build
	docker push $(IMAGE):$(TAG)

dev:
	go run main.go

clean:
	rm -f wpcache
