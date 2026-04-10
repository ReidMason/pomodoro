run-server-test:
  LOOP=true POMODORO_DURATION=5 SHORT_BREAK_DURATION=3 LONG_BREAK_DURATION=5 PORT=8081 go run ./cmd/server
run-tui-test:
  cd clients/tui && go run . -s localhost:8081


run-server:
  docker compose up -d
run-tui:
  cd clients/tui && go run . -s localhost:8080
