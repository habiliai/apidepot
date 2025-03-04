FROM golang:1.22-alpine3.20 AS builder

RUN apk add --no-cache make

WORKDIR /app
COPY . /app
RUN --mount=type=cache,target=/go/pkg\
    echo `go env GOOS` &&\
    echo `go env GOARCH` &&\
    make clean bin/apidepot

FROM alpine:3.20
RUN apk add --no-cache oci-cli

USER 1000:1000
WORKDIR /app
RUN chown -R 1000:1000 /app
ENV PATH="/app/bin:${PATH}" HOME="/app"
COPY --chown=1000:1000 --from=builder /app/bin/apidepot /app/bin/apidepot

ENTRYPOINT ["apidepot"]
