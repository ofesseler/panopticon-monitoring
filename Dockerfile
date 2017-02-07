FROM alpine:3.5
MAINTAINER Oliver Fesseler <ofesseler@gmail.com>

ENV PROM_VER="1.4.1"
ENV PROM_URL="https://github.com/prometheus/prometheus/releases/download"

RUN apk add --update wget git tar nginx supervisor go

RUN mkdir /go
ENV GOPATH /go
RUN go get github.com/looplab/fsm

# Download Prometheus from github
RUN wget -q -O /tmp/prometheus.tar.gz \
        ${PROM_URL}/v${PROM_VER}/prometheus-${PROM_VER}.linux-amd64.tar.gz && \
    tar -xzf /tmp/prometheus.tar.gz -C /tmp

# Install Prometheus from github
RUN mkdir -p /etc/prometheus && \
    mv /tmp/prometheus-${PROM_VER}.linux-amd64/prometheus /bin/ && \
    mv /tmp/prometheus-${PROM_VER}.linux-amd64/promtool /bin/ && \
    mv /tmp/prometheus-${PROM_VER}.linux-amd64/console_libraries/ \
       /etc/prometheus/ && \
    mv /tmp/prometheus-${PROM_VER}.linux-amd64/consoles/ \
       /etc/prometheus/ && \
    rm -rf /tmp/prometheus*

# Nginx config
RUN adduser -D -u 1000 -g 'www' www && \
    mkdir /www && \
    chown -R www:www /var/lib/nginx && \
    chown -R www:www /www
ADD docker-assets/nginx/nginx.conf /etc/nginx/nginx.conf

# Prometheus config
#ADD docker-assets/prometheus.yml /etc/prometheus/prometheus.yml


# Supervisord config
RUN mkdir -p /var/log/supervisord
RUN mkdir -p /var/log/panopticon-dev/
ADD docker-assets/supervisord.conf /etc/supervisor/conf.d/supervisord.conf

# Test-metrics files
#ADD docker-assets/metrics/* /www/

EXPOSE 9090 8080

CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"]
#ENTRYPOINT /bin/sh
