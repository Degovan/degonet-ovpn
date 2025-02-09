FROM debian:latest

WORKDIR /etc/openvpn

# Install Packages
RUN apt update && apt install -y \
    openvpn \
    iptables \
    curl \
    git \
    php-cli \
    php-fpm \
    php-curl \
    php-mysql \
    php-xml \
    php-zip \
    php-phar \
    php-json \
    php-mbstring \
    && rm -rf /var/lib/apt/lists/*

# Install Composer
RUN curl -sS https://getcomposer.org/installer | php -- --install-dir=/usr/local/bin --filename=composer
RUN sed -i 's/;phar.readonly = On/phar.readonly = Off/' /etc/php/8.2/cli/php.ini

RUN mkdir -p /etc/openvpn
COPY . /etc/openvpn

RUN composer install -d /etc/openvpn/auth \
    && composer run build -d /etc/openvpn/auth

RUN chmod +x /etc/openvpn/auth.phar
RUN bash setup-certs.sh

CMD ["openvpn", "--config", "/etc/openvpn/server.conf"]
