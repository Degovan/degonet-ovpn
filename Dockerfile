FROM debian:latest

WORKDIR /etc/openvpn

RUN apt update && apt install -y \
    openvpn \
    iptables \
    php-cli \
    php-fpm \
    php-curl \
    php-mysql \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /etc/openvpn
COPY . /etc/openvpn
RUN chmod +x /etc/openvpn/auth.phar

CMD ["openvpn", "--config", "/etc/openvpn/server.conf"]
