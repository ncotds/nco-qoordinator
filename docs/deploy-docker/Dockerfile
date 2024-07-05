ARG GO_VERSION

FROM golang:${GO_VERSION:-err} as builder

LABEL org.ncoqapi.author="Vitaly Shatskikh vvsh.msk@gmail.com"\
      org.ncoqapi.description="Application builder"

WORKDIR /app

ARG VERSION

ENV GOOS="linux" \
    GOARCH="amd64" \
    CGO_ENABLED=0

COPY go.mod go.sum ./

RUN go mod download -x

COPY pkg ./pkg
COPY cmd ./cmd
COPY internal ./internal

RUN go build -ldflags="-s -w -X='main.version=${VERSION:-err}'" -o ncoq-api ./cmd/ncoq-api/...


FROM scratch as app

LABEL org.ncoqapi.author="Vitaly Shatskikh vvsh.msk@gmail.com" \
      org.ncoqapi.description="Tool that provides HTTP-API to proxy requests to one or more \"OMNIbus Object Server\" instances and collects all results"

COPY config/example.yml /etc/ncoq-api/config.yml
COPY --from=builder /app/ncoq-api /usr/local/bin/ncoq-api

ENV NCOQ_HTTP_LISTEN=":4000"

EXPOSE 4000

USER 1000:1000

ENTRYPOINT [ "ncoq-api", "-c", "/etc/ncoq-api/config.yml" ]
