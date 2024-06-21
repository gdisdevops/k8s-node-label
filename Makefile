PROJECT_NAME := k8s-node-label
GITHUB_PATH := github.com/daspawnw/k8s-node-label

all: clean test bin

clean:
	rm -rf bin

test:
	go test ${GITHUB_PATH}/...

bin: bin/linux bin/darwin

bin/%:
	mkdir -p $@
	CGO_ENABLED=0 GOOS=$(word 1, $(subst /, ,$*)) GOARCH=amd64 go build -o "$@" ${GITHUB_PATH}/cmd/${PROJECT_NAME}/...

run:
	go run ${GITHUB_PATH}/cmd/${PROJECT_NAME}/...

docker:
	docker buildx build -t k8s-node-label:local --load .
