FROM golang:1.18.0 AS builder

WORKDIR /src
COPY . /src

RUN GOPROXY=https://goproxy.cn && make build

FROM buildpack-deps:bullseye-curl

RUN apt-get update && apt-get install -y --no-install-recommends \
		ca-certificates  \
        netbase \
        && rm -rf /var/lib/apt/lists/ \
        && apt-get autoremove -y && apt-get autoclean -y

WORKDIR /app

COPY --from=builder /src/bin /app

EXPOSE 8088

ENTRYPOINT ["./pulsar","http"]