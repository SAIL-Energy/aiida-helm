FROM golang:1.23.5-alpine

LABEL maintainer="eddie.developers@fh-hagenberg.at"

WORKDIR /src

COPY . .

RUN go mod download

RUN go build -o /usr/aiida-installer

EXPOSE 8010

CMD ["/usr/aiida-installer"]