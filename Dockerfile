FROM golang AS builder


WORKDIR /home/kingdo/GolandProjects/src/github.com/Kingdo777/serverless.instance.select/
ADD  . .

ENV GOPATH  /home/kingdo/GolandProjects/

RUN CGO_ENABLED=0 go build -o  ./cmd/measure/app ./cmd/measure/

FROM gcr.io/distroless/base

EXPOSE 8081
COPY  ./cmd/measure/app /app

ENTRYPOINT ["/app"]