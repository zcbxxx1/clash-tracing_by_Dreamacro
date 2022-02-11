package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/avast/retry-go/v4"
)

func handleReport(addr, clashHost, clashToken string) {
	dialer := net.Dialer{}
	for {
		var conn net.Conn
		retry.Do(
			func() (err error) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()
				conn, err = dialer.DialContext(ctx, "tcp", addr)
				return
			},
			retry.Attempts(0),
			retry.DelayType(func(n uint, err error, config *retry.Config) time.Duration {
				max := time.Duration(n)
				if max > 8 {
					max = 8
				}
				duration := time.Second * max * max
				fmt.Printf("dial %s failed %d times: %v, wait %s\n", addr, n, err, duration.String())
				return duration
			}),
			retry.MaxDelay(time.Second*64),
		)

		fmt.Printf("dial %s success, send data to vector\n", addr)
		handleTCPConn(conn, clashHost, clashToken)

		conn.Close()
	}
}

func handleTCPConn(conn net.Conn, clashHost string, clashToken string) {
	trafficCh := make(chan []byte)
	tracingCh := make(chan []byte)

	ctx, cancel := context.WithCancel(context.Background())

	trafficDone := dialTrafficChan(ctx, clashHost, clashToken, trafficCh)
	tracingDone := dialTracingChan(ctx, clashHost, clashToken, tracingCh)

Out:
	for {
		var buf []byte
		select {
		case buf = <-trafficCh:
		case buf = <-tracingCh:
		}
		if _, err := conn.Write(buf); err != nil {
			break Out
		}
	}

	cancel()

	go func() {
		for range trafficCh {
		}
		for range tracingCh {
		}
	}()

	<-trafficDone
	<-tracingDone
}

func dialTrafficChan(ctx context.Context, clashHost string, clashToken string, ch chan []byte) chan struct{} {
	var clashUrl string
	if clashToken == "" {
		clashUrl = fmt.Sprintf("ws://%s/traffic", clashHost)
	} else {
		clashUrl = fmt.Sprintf("ws://%s/traffic?token=%s", clashHost, clashToken)
	}

	return dialWebsocketToChan(context.Background(), clashUrl, ch)
}

func dialTracingChan(ctx context.Context, clashHost string, clashToken string, ch chan []byte) chan struct{} {
	var clashUrl string
	if clashToken == "" {
		clashUrl = fmt.Sprintf("ws://%s/profile/tracing", clashHost)
	} else {
		clashUrl = fmt.Sprintf("ws://%s/profile/tracing?token=%s", clashHost, clashToken)
	}

	return dialWebsocketToChan(context.Background(), clashUrl, ch)
}
