package main

import (
	"os"
	"os/signal"
	"syscall"
)

func envOrDefault(env string, def string) string {
	value, exist := os.LookupEnv(env)
	if !exist {
		return def
	}
	return value
}

func main() {
	clashHost := envOrDefault("CLASH_HOST", "localhost:9090")
	clashToken := envOrDefault("CLASH_TOKEN", "")
	vectorAddr := envOrDefault("VECTOR_ADDR", "localhost:9000")

	go handleReport(vectorAddr, clashHost, clashToken)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}
