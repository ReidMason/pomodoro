FROM golang:1.26.1 AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 go build -o /pomodoro-server ./cmd/server

FROM scratch

COPY --from=builder /pomodoro-server ./pomodoro-server

ENTRYPOINT ["./pomodoro-server"]
