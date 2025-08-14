FROM golang:alpine3.20 as builder

ENV GOOS="linux"
ENV CGO_ENABLED="0"
ARG GOARCH="amd64"

ARG CMD="http"

WORKDIR /app

COPY cmd/${CMD} cmd/${CMD}
COPY pb pb
COPY core core

COPY go.mod go.mod
COPY go.sum go.sum

RUN go build -o main ./cmd/${CMD}/main.go

EXPOSE 8000

FROM alpine:3.20 as prod
COPY --from=builder /app/main /bin/
ENTRYPOINT  ["/bin/main"]

FROM alpine:3.20 as static
COPY --from=builder /app/main /bin/
ARG CMD
COPY cmd/${CMD}/static /static
ENTRYPOINT  ["/bin/main"]


