FROM golang:1.7
MAINTAINER Oliver Fesseler <oliver@fesseler.info>

EXPOSE 8888

ADD . $GOPATH/src/github.com/panopticon/panopticon
COPY static/ static/

RUN go install github.com/panopticon/panopticon

ENTRYPOINT $GOPATH/bin/panopticon -prom-host zwuggl:9090
