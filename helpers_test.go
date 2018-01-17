package wexapi

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
)

// SetHTTPClient sets http client for the client.
func SetHTTPClient(httpClient *http.Client) Option {
	return func(cli *Client) {
		cli.httpClient = httpClient
	}
}

func createFakeServer(h http.Handler) *httptest.Server {
	server := httptest.NewTLSServer(h)

	return server
}

func transportForTesting(server *httptest.Server) *http.Transport {
	return &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			return net.Dial("tcp", server.URL[strings.LastIndex(server.URL, "/")+1:])
		},
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
}

func testingHTTPClient(server *httptest.Server) *http.Client {
	return &http.Client{
		Transport: transportForTesting(server),
	}
}

func compareAsStrings(got, expect interface{}) bool {
	return fmt.Sprintf("%s", got) == fmt.Sprintf("%s", expect)
}
