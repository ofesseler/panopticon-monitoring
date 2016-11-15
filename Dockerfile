FROM golang:1.7

ADD . $GOPATH/src/fesseler.info/salus/salus
COPY static/ static/

RUN go install fesseler.info/salus/salus

ENTRYPOINT $GOPATH/bin/salus

EXPOSE 8888

