FROM golang:1.25-alpine AS builder

WORKDIR /build

COPY auth-src/go.mod auth-src/go.sum* ./
RUN go mod download

COPY auth-src/ ./
RUN go mod tidy && CGO_ENABLED=0 go build -ldflags="-s -w" -o auth .

FROM alpine:latest

RUN apk add --no-cache \
    openvpn \
    iptables \
    bash \
    iputils \
    curl \
    openssl

WORKDIR /etc/openvpn

COPY --from=builder /build/auth /etc/openvpn/auth
COPY .docker/entrypoint.sh /etc/openvpn/
COPY server.conf /etc/openvpn/
COPY setup-certs.sh /etc/openvpn/

ENV DB_FILE=/etc/openvpn/data/users.sqlite
ENV DEFAULT_NETMASK=255.255.255.0
ENV CCD_DIR=/etc/openvpn/ccds

RUN chmod +x /etc/openvpn/auth \
    && chmod +x /etc/openvpn/entrypoint.sh \
    && mkdir -p /etc/openvpn/ccds /etc/openvpn/data

RUN bash /etc/openvpn/setup-certs.sh

CMD ["/etc/openvpn/entrypoint.sh"]
