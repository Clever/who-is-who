# who-is-who api service
FROM golang:1.4

ENV service "who-is-who"
ENV dir "/go/src/github.com/Clever/$service"

RUN mkdir -p "$dir"
ADD . "$dir"
WORKDIR "$dir"

RUN go get ./...
RUN go build

CMD ["./$service"]
