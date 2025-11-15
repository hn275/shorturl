ARG GO_VERSION=1
FROM golang:1.25.1-bookworm AS builder

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -v -o /run-app .


FROM debian:bookworm

COPY --from=builder /run-app /app
COPY /public /public
COPY /assets /assets
CMD ["/app"]
