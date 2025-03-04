FROM golang:1.23-alpine

ENV CGO_ENABLED=1

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN apk add --no-cache gcc g++ git openssh-client
RUN go build -o app

EXPOSE 5080

CMD ["/build/app"]