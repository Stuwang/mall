FROM golang:1.24-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG BIN=biz
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/${BIN}

FROM alpine:3.21

RUN adduser -D -u 65532 app
COPY --from=build /out/app /usr/local/bin/app
USER 65532

EXPOSE 8000 8081 9000
ENTRYPOINT ["/usr/local/bin/app"]
