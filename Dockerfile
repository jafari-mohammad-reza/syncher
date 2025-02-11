FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /server ./cmd/server.go
RUN go build -o /client ./cmd/client.go
FROM alpine:latest as run
WORKDIR /app
COPY --from=build /server /server
COPY --from=build /app/server.yaml /app
RUN chmod a+x /server
CMD ["/server"]
