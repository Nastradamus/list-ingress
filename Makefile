.DEFAULT_GOAL = all

# Build the project
all: build

build:
	docker image build -t list-ingress:latest .

run-local:
	docker run --rm -v $(HOME)/.kube/config:/root/.kube/config -p 8080:8080 list-ingress:latest list-ingress -run-outside-cluster -v 1

.PHONY: all build run-local
