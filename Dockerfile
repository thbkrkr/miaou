FROM golang:1.10-alpine3.7 as builder
ENV GOBIN=/go/bin/ GOPATH=/go
WORKDIR /go/src/github.com/thbkrkr/miaou
COPY . /go/src/github.com/thbkrkr/miaou
RUN CGO_ENABLED=0 GOOS=linux go build

FROM alpine:3.7
RUN apk --no-cache add ca-certificates
COPY --from=builder \
  /go/src/github.com/thbkrkr/miaou/miaou \
  /usr/local/bin/miaou
ENTRYPOINT ["miaou"]