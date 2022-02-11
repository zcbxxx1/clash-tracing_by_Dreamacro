package main

import (
	"context"
	"sync"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/gorilla/websocket"
)

type Conn struct {
	conn *websocket.Conn
	once sync.Once
	Ch   chan []byte
}

func (c *Conn) Close() error {
	c.once.Do(func() {
		close(c.Ch)
	})
	return c.conn.Close()
}

func dialWebsocketToChan(ctx context.Context, url string, ch chan []byte) chan struct{} {
	done := make(chan struct{}, 1)
	go func() {
		for {
			var conn *Conn
			retry.Do(
				func() (err error) {
					conn, err = dialWebsocket(ctx, url)
					return
				},
				retry.Attempts(0),
				retry.DelayType(func(n uint, err error, config *retry.Config) time.Duration {
					println("dial websocket failed", url, n, err.Error())
					return retry.BackOffDelay(n, err, config)
				}),
				retry.RetryIf(func(e error) bool {
					return e != context.Canceled
				}),
				retry.MaxDelay(time.Second*64),
			)
		Out:
			for {
				select {
				case <-ctx.Done():
					done <- struct{}{}
					conn.Close()
					return
				case buf, open := <-conn.Ch:
					if !open {
						conn.Close()
						break Out
					}
					ch <- buf
				}
			}

		}
	}()
	return done
}

func dialWebsocket(ctx context.Context, url string) (*Conn, error) {
	dialer := &websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	conn, _, err := dialer.DialContext(ctx, url, nil)
	if err != nil {
		return nil, err
	}

	// ping pong
	go func() {
		ticker := time.NewTicker(time.Second * 60)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				conn.WriteMessage(websocket.PingMessage, nil)
			}
		}
	}()

	ch := make(chan []byte)
	c := &Conn{conn: conn, Ch: ch}

	go func() {
		defer c.Close()
	Loop:
		for {
			select {
			case <-ctx.Done():
				conn.Close()
				return
			default:
				_, buf, err := conn.ReadMessage()
				if err != nil {
					break Loop
				}
				ch <- buf
			}
		}
	}()

	return c, nil
}
