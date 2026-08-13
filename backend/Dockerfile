# --- build stage ---
# --- build stage ---
FROM golang:1.25-alpine AS build

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /yven-server ./cmd/server

# --- run stage ---
FROM alpine:3.19

RUN apk add --no-cache ca-certificates

COPY --from=build /yven-server /usr/local/bin/yven-server

EXPOSE 8080

ENTRYPOINT ["yven-server"]
