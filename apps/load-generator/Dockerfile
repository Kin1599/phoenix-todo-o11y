# Этап сборки
FROM golang:1.24-alpine AS build

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/load-ui .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=build /app/load-ui .

CMD ["./load-ui"]
