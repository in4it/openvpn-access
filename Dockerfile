#
# Build go project
#
FROM golang:1.12-alpine as go-builder

WORKDIR /go/src/github.com/in4it/openvpn-access

COPY . .

RUN apk add -u -t build-tools curl git && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o openvpn-access cmd/server/main.go


#
# Runtime container
#
FROM alpine:latest  

RUN apk --no-cache add ca-certificates && mkdir -p /app

WORKDIR /app

COPY --from=go-builder /go/src/github.com/in4it/openvpn-access/openvpn-access /app/openvpn-access

CMD ["./openvpn-access"]  

