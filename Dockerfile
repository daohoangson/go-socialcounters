FROM golang:1.11.11-alpine3.10

RUN apk add --update build-base git

COPY go.mod /src/
COPY go.sum /src/
WORKDIR /src
RUN go mod download

COPY . /src

EXPOSE 8080

CMD ["go", "run", "main.go"]
