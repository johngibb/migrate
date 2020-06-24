FROM golang:1.14

RUN mkdir -p /go/src/github.com/johngibb/migrate
WORKDIR /go/src/github.com/johngibb/migrate
COPY . .

RUN go install ./...
