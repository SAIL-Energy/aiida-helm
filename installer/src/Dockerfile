FROM golang:1.23.5-alpine AS builder

WORKDIR /src

COPY . .

RUN go mod download

RUN go build -o /usr/aiida-installer

FROM alpine:3.21.2

LABEL maintainer="eddie.developers@fh-hagenberg.at"

WORKDIR /src

COPY --from=builder /usr/aiida-installer /usr/aiida-installer

EXPOSE 8010

CMD ["/usr/aiida-installer"]