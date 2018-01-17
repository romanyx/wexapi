package wexapi

import (
	"net/http"
	"strconv"
	"time"
)

const (
	defaultTimeout = 10 * time.Second
)

// Option for initializer.
type Option func(*Client)

// SetTimeout sets timeout for the http client.
func SetTimeout(timeout time.Duration) Option {
	return func(cli *Client) {
		cli.httpClient.Timeout = timeout
	}
}

// Client for requesting wex api.
// Use NewClient to initialize one.
type Client struct {
	key, secret string
	httpClient  *http.Client

	noncePool chan int64
}

// NewClient returns initialized client.
func NewClient(key, secret string, options ...Option) *Client {
	cli := Client{
		key:    key,
		secret: secret,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		noncePool: make(chan int64),
	}

	go func() {
		cli.noncePool <- time.Now().Unix()
	}()

	for _, option := range options {
		option(&cli)
	}

	return &cli
}

func (cli *Client) nonce() string {
	nonce := <-cli.noncePool
	go func() {
		cli.noncePool <- nonce + 1
	}()
	return strconv.FormatInt(nonce, 10)
}
