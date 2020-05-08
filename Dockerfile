# 1) Prepare and build deps (GO111MODULE version)
FROM golang:1.13.5-alpine3.11 as download-deps
ENV GO111MODULE=on
RUN mkdir /MultiStage
WORKDIR /MultiStage
COPY go.mod .
COPY go.sum .
# Download and save deps
RUN go mod download
COPY . .
# 2) Build app
FROM golang:1.13.5-alpine3.11 as build-deps
RUN mkdir /MultiStage
WORKDIR /MultiStage
COPY --from=download-deps /MultiStage/ /MultiStage/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /list-ingress
COPY . .
# 3) Run pre-builded app in alpine
FROM alpine:3.11.3
RUN mkdir /root/.kube
COPY --from=build-deps /MultiStage/list-ingress /bin/list-ingress
CMD ["/bin/list-ingress"]
