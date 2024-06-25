ARG GO_VERSION

FROM golang:${GO_VERSION:-err} as builder

LABEL org.ncoqapi.author="Vitaly Shatskikh vvsh.msk@gmail.com"\
      org.ncoqapi.description="Application builder"

RUN apt-get update \
   && apt-get install -y --no-install-recommends \
      freetds-dev \
   && rm -rf /var/lib/apt/lists/*

WORKDIR /app

ARG VERSION

ENV GOOS="linux" \
    GOARCH="amd64"

COPY go.mod go.sum ./

RUN go mod download -x

COPY pkg ./pkg
COPY cmd ./cmd
COPY internal ./internal

RUN go build -ldflags="-s -w -X='main.version=${VERSION:-err}'" -o ncoq-api ./cmd/ncoq-api/...


FROM alpine:3.18 as app

LABEL org.ncoqapi.author="Vitaly Shatskikh vvsh.msk@gmail.com" \
      org.ncoqapi.description="Tool that provides HTTP-API to proxy requests to one or more \"OMNIbus Object Server\" instances and collects all results"

RUN apk add --no-cache --update libc6-compat freetds-dev

COPY config/example.yml /etc/ncoq-api/config.yml
COPY --from=builder /app/ncoq-api /usr/local/bin/ncoq-api

ENV NCOQ_HTTP_LISTEN=":4000"

EXPOSE 4000

USER 1000:1000

ENTRYPOINT [ "ncoq-api", "-c", "/etc/ncoq-api/config.yml" ]
