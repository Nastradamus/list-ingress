# 1) Download dep modules
FROM golang:1.13.5-alpine3.11 as build-env

RUN mkdir /MultiStage
WORKDIR /MultiStage
COPY go.mod .
COPY go.sum .

COPY list-ingress.go .

# Build the binary
ENV GO111MODULE=on
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /MultiStage/list-ingress

# 2) Copy builded binary
FROM alpine:3.11.3

COPY --from=build-env /MultiStage/list-ingress /bin/list-ingress

RUN mkdir /root/.kube

CMD ["/bin/list-ingress"]
