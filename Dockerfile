FROM golang:alpine as builder

WORKDIR /go/src/github.com/seguidor777/portfel

COPY cmd cmd
COPY internal internal
COPY go.mod go.sum ./

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /bin/portfel github.com/seguidor777/portfel/cmd/portfel
RUN mkdir -p /usr/share/portfel

FROM scratch

WORKDIR /app

COPY --from=builder /bin/portfel /bin/portfel
COPY --from=builder /usr/share/portfel /usr/share/portfel
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["portfel"]
