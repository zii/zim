FROM golang:alpine AS builder
ARG APP
ENV GOPROXY=https://proxy.golang.com.cn,direct
COPY ./ /go/src/
WORKDIR /go/src
RUN set -eux \
    && if [ "${APP}" = "zimapi" ]; then go build -o ./zimapi zim.cn/service/apisvc/cmd; fi \
    && if [ "${APP}" = "zimpush" ]; then go build -o ./zimpush zim.cn/service/pushsvc/cmd; fi \
    && if [ "${APP}" = "zimcron" ]; then go build -o ./zimcron zim.cn/service/cronsvc/cmd; fi

FROM alpine AS app-image
ARG APP
ARG PORT
ENV APP_NAME=${APP}
ENV APP_PORT=${PORT}
EXPOSE ${PORT}
RUN set -eux \
    && sed -i 's@dl-cdn.alpinelinux.org@mirrors.aliyun.com@g' /etc/apk/repositories \
    && apk add --no-cache tzdata curl \
    && ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && addgroup -g 1000 -S user \
    && adduser -S -D -u 1000 -s /sbin/nologin -G user -g user user
COPY --from=builder /go/src/${APP_NAME} /usr/local/bin/
COPY --from=builder /go/src/bin/dev/config_86.toml /home/user/config.toml
COPY docker-entrypoint.sh /usr/local/bin/
HEALTHCHECK --interval=5s --timeout=3s \
  CMD curl -fs http://localhost:${APP_PORT}/health || exit 1
WORKDIR /home/user
USER user
ENTRYPOINT ["docker-entrypoint.sh"]
