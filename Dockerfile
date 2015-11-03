#
# Eventhon Dockerfile
#

FROM golang:1.5.1

MAINTAINER Lucien Zagabe <rz@ognitio.com>

WORKDIR /go/src/ognitio.com/rzagabe/eventhon
ADD . /go/src/ognitio.com/rzagabe/eventhon
RUN go get ./... && go install

CMD ["echo done"]
