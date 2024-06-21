#
# Build go project
#
FROM golang:1.19-alpine as go-builder

WORKDIR /openvpn-access

COPY . .

RUN apk add -u -t build-tools curl git && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o openvpn-access cmd/server/main.go


#
# Runtime container
#
FROM alpine:3.17 

RUN apk --no-cache add ca-certificates && mkdir -p /app

WORKDIR /app

COPY --from=go-builder /openvpn-access/openvpn-access /app/openvpn-access

CMD ["./openvpn-access"]  

