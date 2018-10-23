FROM golang:1.11 AS builder

# Copy the code from the host and compile it
WORKDIR /build
COPY . ./
RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o /app .

FROM alpine:latest as certs
RUN apk --update add ca-certificates

FROM scratch
COPY --from=builder /app ./
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
ENTRYPOINT ["./app"]

