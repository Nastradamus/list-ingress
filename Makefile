CI_SCRIPTS_PATH = ./ci

.DEFAULT_GOAL = all
.PHONY: all build run-local

# Build the project
all: build

build:
	docker image build -t list-ingress:latest .

run-local:
	docker run --rm -v $(HOME)/.kube/config:/root/.kube/config -p 8080:8080 list-ingress:latest list-ingress -run-outside-cluster -v 1

lint:
	${CI_SCRIPTS_PATH}/linters.sh

run-local-outside:
	go build && ./list-ingress -run-outside-cluster klog -v 1

