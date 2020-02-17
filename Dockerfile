FROM golang:1.13 as builder

WORKDIR /go/src/app
COPY . .

RUN CGO_ENABLED=0 go install -v ./...

FROM scratch as solar

# COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/solar-battery-monitoring /go/bin/solar-battery-monitoring
# COPY --from=builder --chown=yinyo:0 /tmp /tmp

CMD ["/go/bin/solar-battery-monitoring"]