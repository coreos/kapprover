FROM       golang:alpine
ADD        . /go/src/github.com/coreos/kapprover
RUN        go install github.com/coreos/kapprover/cmd/kapprover

FROM alpine
MAINTAINER Quentin Machu <quentin.machu@coreos.com>
COPY --from=0 /go/bin/kapprover .
ENTRYPOINT ["/kapprover"]
