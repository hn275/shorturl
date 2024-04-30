FROM golang:1.22.2-bullseye AS build
WORKDIR /shorturl
COPY . .
RUN go mod tidy
RUN go build -o bin/shorturl main.go

FROM debian:bullseye
WORKDIR /shorturl
COPY --from=build /shorturl/bin/shorturl /bin/shorturl
CMD ["/bin/shorturl"]
