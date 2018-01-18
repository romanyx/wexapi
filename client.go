package wexapi

import (
	"errors"
	"math"
	"net/http"
	"strconv"
	"time"
)

const (
	defaultTimeout = 10 * time.Second
)

// ErrNonceOverflow caused when nonce is equal to 4294967294
// which is maximum size for api key at wex.
var ErrNonceOverflow = errors.New("max value reached: create new key")

// Option for initializer.
type Option func(*Client)

// SetHTTPClient sets http client for the client.
func SetHTTPClient(httpClient *http.Client) Option {
	return func(cli *Client) {
		cli.httpClient = httpClient
	}
}

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

	noncePool chan uint32 // max is 4294967294
}

// NewClient returns initialized client.
func NewClient(key, secret string, options ...Option) *Client {
	cli := Client{
		key:    key,
		secret: secret,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		noncePool: make(chan uint32),
	}

	go func() {
		cli.noncePool <- uint32(time.Now().Unix())
	}()

	for _, option := range options {
		option(&cli)
	}

	return &cli
}

func (cli *Client) nonce() (string, error) {
	nonce := <-cli.noncePool
	if nonce == uint32(math.MaxUint32)-1 {
		go func() {
			cli.noncePool <- nonce
		}()
		return "", ErrNonceOverflow
	}
	go func() {
		cli.noncePool <- nonce + 1
	}()
	return strconv.FormatUint(uint64(nonce), 10), nil
}
