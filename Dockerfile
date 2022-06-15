# syntax = docker/dockerfile:1.2
# WARNING! 使用 `DOCKER_BUILDKIT=1` 和 `docker build` 來啟用 --mount 功能。

## 準備基礎鏡像。
#
FROM golang:1.18.0-bullseye as base

RUN apt update && \
    apt-get install -y \
         build-essential \
         ca-certificates \
         curl

# 啟用更快的模塊下載。
ENV GOPROXY https://proxy.golang.org

## 建設者階段。
#
FROM base as builder

WORKDIR /ignite

# 緩存依賴項。
COPY ./go.mod . 
COPY ./go.sum . 
RUN go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build go install -v ./...

## 準備最終圖像。
#
FROM base

RUN useradd -ms /bin/bash tendermint
USER tendermint

COPY --from=builder /go/bin/ignite /usr/bin

WORKDIR /apps

# 請參閱暴露端口的文檔：
#   https://docs.ignite.com/kb/config.html#host
EXPOSE 26657
EXPOSE 26656
EXPOSE 6060 
EXPOSE 9090 
EXPOSE 1317 

ENTRYPOINT ["ignite"]
