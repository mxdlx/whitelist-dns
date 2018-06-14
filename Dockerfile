FROM golang:1.10-alpine
MAINTAINER aaraujo@protonmail.ch

WORKDIR /go/src/whitelist-dns
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["whitelist-dns"]
