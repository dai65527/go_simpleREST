FROM golang:1.16.4-buster AS builder
RUN mkdir /workdir
WORKDIR /workdir
COPY ./srcs/main.go .
COPY ./srcs/go.mod .
ENV GORARCH amd64
ENV GOOS linux
RUN go mod tidy
RUN go build -o app .

FROM debian:buster
COPY --from=builder /workdir/app .
CMD ["./app"]
