FROM golang:1.10

RUN mkdir -p /go/src/github.com/johngibb/migrate
WORKDIR /go/src/github.com/johngibb/migrate
COPY . .

RUN go install ./...
