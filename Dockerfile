FROM golang:1.11-stretch as builder
WORKDIR /go/src/github.com/moov-io/auth
RUN apt-get update && apt-get install make gcc g++
RUN adduser -q --gecos '' --disabled-login --shell /bin/false moov
COPY . .
RUN make build
USER moov

FROM debian:9
RUN apt-get update && apt-get install -y ca-certificates
COPY --from=builder /go/src/github.com/moov-io/auth/bin/auth /bin/auth
COPY --from=builder /etc/passwd /etc/passwd
USER moov
EXPOSE 8080
EXPOSE 9090
VOLUME "/data"
ENTRYPOINT ["/bin/auth"]
