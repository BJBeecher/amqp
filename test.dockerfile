FROM golang:latest

RUN mkdir /app

ADD . /app/

WORKDIR /app

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o main .

CMD ["go", "test", "-v", "./test"]