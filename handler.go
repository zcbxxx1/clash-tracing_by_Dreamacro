package main

import (
	"context"
	"fmt"
	"net"
	"time"
)

func handleReport(addr, clashHost, clashToken string) {
	dialer := net.Dialer{}
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		conn, err := dialer.DialContext(ctx, "tcp", addr)
		cancel()
		if err != nil {
			fmt.Printf("dial %s failed: %v, 5s later retry\n", addr, err)
			time.Sleep(time.Second * 5)
			continue
		}
		fmt.Printf("dial %s success, send data to vector\n", addr)
		handleTCPConn(conn, clashHost, clashToken)
	}
}

func handleTCPConn(conn net.Conn, clashHost string, clashToken string) {
	trafficCh := dialTrafficChan(context.Background(), clashHost, clashToken)
	tracingCh := dialTracingChan(context.Background(), clashHost, clashToken)

	for {
		var buf []byte
		select {
		case buf = <-trafficCh:
		case buf = <-tracingCh:
		}
		if _, err := conn.Write(buf); err != nil {
			return
		}
	}
}

func dialTrafficChan(ctx context.Context, clashHost string, clashToken string) chan []byte {
	var clashUrl string
	if clashToken == "" {
		clashUrl = fmt.Sprintf("ws://%s/traffic", clashHost)
	} else {
		clashUrl = fmt.Sprintf("ws://%s/traffic?token=%s", clashHost, clashToken)
	}

	return dialWebsocketChan(context.Background(), clashUrl)
}

func dialTracingChan(ctx context.Context, clashHost string, clashToken string) chan []byte {
	var clashUrl string
	if clashToken == "" {
		clashUrl = fmt.Sprintf("ws://%s/profile/tracing", clashHost)
	} else {
		clashUrl = fmt.Sprintf("ws://%s/profile/tracing?token=%s", clashHost, clashToken)
	}

	return dialWebsocketChan(context.Background(), clashUrl)
}
