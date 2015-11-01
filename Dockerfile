#
# Eventos Dockerfile
#

FROM golang:1.5.1

MAINTAINER Lucien Zagabe <rz@ognitio.com>

WORKDIR /go/src/ognitio.com/rzagabe/eventos
ADD . /go/src/ognitio.com/rzagabe/eventos
RUN go get ./... && go install

CMD ["echo done"]
