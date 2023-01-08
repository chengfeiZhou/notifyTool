FROM golang:alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0
# http://10.11.12.35:32888 为自建的go第三方库代理
ENV GOPROXY http://10.11.12.35:32888,https://goproxy.cn,direct
RUN echo -e http://mirrors.ustc.edu.cn/alpine/v3.15/main/ > /etc/apk/repositories
RUN apk update  && apk add tzdata

WORKDIR /build

ADD go.mod .
ADD go.sum .
RUN go mod download
COPY . .
RUN go build -ldflags="-s -w" -o /app/security_notify cmd/security_notify/security_notify.go


FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /usr/share/zoneinfo/Asia/Shanghai
ENV TZ Asia/Shanghai

WORKDIR /app
COPY --from=builder /app/security_notify /app/security_notify

EXPOSE 8900

ENTRYPOINT ["./security_notify"]
CMD ["-addr", "8900"]