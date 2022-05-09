FROM golang:1.18

ENV GORPOXY https://goproxy.cn,direct
RUN go install github.com/go-delve/delve/cmd/dlv@v1.8.3 && go install github.com/cosmtrek/air@latest

ENV GO111MODULE on

ENV CONFIG_ETCD "etcd://etcd:2379"
ENV BUILD_MAIN_FILE "cmd/main.go"
ENV BUILD_MAIN_FLAG "http"

VOLUME /app
WORKDIR /app

COPY conf/.air.toml /tmp/.air.toml
COPY conf/dev.sh /tmp/dev.sh

ENTRYPOINT ["air","-c","/tmp/.air.toml"]

