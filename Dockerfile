FROM       golang:alpine
MAINTAINER Quentin Machu <quentin.machu@coreos.com>
ADD        . /go/src/github.com/coreos-inc/kapprover
RUN        go install github.com/coreos-inc/kapprover/cmd/kapprover
ENTRYPOINT ["/go/bin/kapprover"]