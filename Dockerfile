FROM --platform=$BUILDPLATFORM golang:1.19.3 as builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /go/src/github.com/seguidor777/portfel

COPY go.mod go.sum ./
COPY cmd cmd
COPY internal internal

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} CGO_ENABLED=0 go build -ldflags="-w -s" -o /bin/portfel github.com/seguidor777/portfel/cmd/portfel
RUN mkdir -p /usr/share/portfel

FROM scratch

WORKDIR /app

COPY --from=builder /bin/portfel /bin/portfel
COPY --from=builder /usr/share/portfel /usr/share/portfel
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["portfel"]
