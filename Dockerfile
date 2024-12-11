FROM golang:1.22.5 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o final-project ./cmd/main.go

FROM ubuntu:latest

WORKDIR /app

COPY --from=builder /app/final-project /app/final-project

COPY web /app/web

COPY internal/database/scheduler.db /app/scheduler.db

COPY .env /app/.env

ENV TODO_PORT=7540
ENV TODO_DBFILE=/app/scheduler.db
ENV TODO_PASSWORD="123"

EXPOSE 7540

CMD ["/app/final-project"]

#docker build -t final-project:v1 .
#docker run -d -p 7540:7540 -v $(pwd)/internal/database/scheduler.db:/app/scheduler.db final-project:v1