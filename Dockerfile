FROM golang:1.24-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o server ./cmd/server

FROM alpine:3.20
RUN apk --no-cache add ca-certificates
COPY --from=build /app/server /server
COPY --from=build /app/migrations /migrations
ENV MIGRATIONS_PATH=/migrations
EXPOSE 8080
CMD ["/server"]
