package wexapi

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

const (
	invalidMethodResponse = `{"success":0, "error":"Invalid method"}`
	infoResponse          = `{
		"server_time":1370814956,
		"pairs":{
			"btc_usd":{
				"decimal_places":3,
				"min_price":0.1,
				"max_price":400,
				"min_amount":0.01,
				"hidden":0,
				"fee":0.2
			}
		}
	}`
	tickerResponse = `{
		"btc_usd":{
			"high":109.88,
			"low":91.14,
			"avg":100.51,
			"vol":1632898.2249,
			"vol_cur":16541.51969,
			"last":101.773,
			"buy":101.9,
			"sell":101.773,
			"updated":1370816308
		}
	}`
	depthResponse = `{
		"btc_usd":{
			"asks":[
				[103.426,0.01]
			],
			"bids":[
				[103.2,2.48502251]
			]
		}
	}`
	tradesResponse = `{
		"btc_usd":[
			{
				"type":"ask",
				"price":103.6,
				"amount":0.101,
				"tid":4861261,
				"timestamp":1370818007
			}
		]
	}`
)

func TestClient_publicRequest(t *testing.T) {
	tests := []struct {
		name    string
		handler http.Handler
		wantErr bool
		errText string
	}{
		{
			name: "error response",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, invalidMethodResponse)
			}),
			wantErr: true,
			errText: "server respond with error: Invalid method",
		},
		{
			name: "error code",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}),
			wantErr: true,
			errText: "server respond with status code 500",
		},
		{
			name: "invalid json",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, "/")
			}),
			wantErr: true,
			errText: "unmarshal to base response: invalid character '/' looking for beginning of value",
		},
		{
			name: "valid json",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, infoResponse)
			}),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createFakeServer(tt.handler)
			defer server.Close()
			httpClient := testingHTTPClient(server)

			cli := NewClient("", "", SetHTTPClient(httpClient))

			err := cli.publicRequest(&baseResponse{}, "any", nil)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Client.publicRequest() error = %s, wantErr %t", err, tt.wantErr)
				}

				if tt.errText != err.Error() {
					t.Errorf("expected err text: %s, got: %s", tt.errText, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Client.publicRequest() error = %s, wantErr %t", err, tt.wantErr)
			}
		})
	}
}

func TestClient_Info(t *testing.T) {
	tests := []struct {
		name    string
		handler http.Handler
		want    InfoResponse
		wantErr bool
	}{
		{
			name: "valid json",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, infoResponse)
			}),
			want: InfoResponse{
				ServerTime: unixTimestamp(time.Unix(1370814956, 0)),
				Pairs: map[string]PairInfo{
					"btc_usd": PairInfo{
						DecimalPlaces: 3,
						MinPrice:      decimal.NewFromFloatWithExponent(0.1, -1),
						MaxPrice:      decimal.NewFromFloatWithExponent(400, 0),
						MinAmount:     decimal.NewFromFloatWithExponent(0.01, -2),
						Fee:           decimal.NewFromFloatWithExponent(0.2, -1),
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createFakeServer(tt.handler)
			defer server.Close()
			httpClient := testingHTTPClient(server)

			cli := NewClient("", "", SetHTTPClient(httpClient))
			got, err := cli.Info()
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Info() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.Info() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Ticker(t *testing.T) {
	tests := []struct {
		name    string
		handler http.Handler
		want    Market
		wantErr bool
	}{
		{
			name: "valid json",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, tickerResponse)
			}),
			want: Market{
				High:             decimal.NewFromFloatWithExponent(109.88, -2),
				Low:              decimal.NewFromFloatWithExponent(91.14, -2),
				Average:          decimal.NewFromFloatWithExponent(100.51, -2),
				Volume:           decimal.NewFromFloatWithExponent(1632898.2249, -4),
				VolumeInCurrency: decimal.NewFromFloatWithExponent(16541.51969, -5),
				Last:             decimal.NewFromFloatWithExponent(101.773, -3),
				Buy:              decimal.NewFromFloatWithExponent(101.9, -1),
				Sell:             decimal.NewFromFloatWithExponent(101.773, -3),
				Updated:          unixTimestamp(time.Unix(1370816308, 0)),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createFakeServer(tt.handler)
			defer server.Close()
			httpClient := testingHTTPClient(server)

			cli := NewClient("", "", SetHTTPClient(httpClient))
			got, err := cli.Ticker("btc_usd")
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Ticker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.Ticker() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Depth(t *testing.T) {
	tests := []struct {
		name    string
		handler http.Handler
		want    OrderBook
		wantErr bool
	}{
		{
			name: "valid json",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, depthResponse)
			}),
			want: OrderBook{
				Asks: []Order{
					Order{
						Rate:   decimal.NewFromFloatWithExponent(103.426, -3),
						Amount: decimal.NewFromFloatWithExponent(0.01, -2),
						Total:  decimal.NewFromFloatWithExponent(1.03426, -5),
					},
				},
				Bids: []Order{
					Order{
						Rate:   decimal.NewFromFloatWithExponent(103.2, -1),
						Amount: decimal.NewFromFloatWithExponent(2.48502251, -8),
						Total:  decimal.NewFromFloatWithExponent(256.454323032, -9),
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createFakeServer(tt.handler)
			defer server.Close()
			httpClient := testingHTTPClient(server)

			cli := NewClient("", "", SetHTTPClient(httpClient))
			got, err := cli.Depth("btc_usd", 1)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Depth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.Depth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Trades(t *testing.T) {
	tests := []struct {
		name    string
		handler http.Handler
		want    []Trade
		wantErr bool
	}{
		{
			name: "valid json",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, tradesResponse)
			}),
			want: []Trade{
				Trade{
					ID:        4861261,
					Type:      "ask",
					Rate:      decimal.NewFromFloatWithExponent(103.6, -1),
					Amount:    decimal.NewFromFloatWithExponent(0.101, -3),
					Timestamp: unixTimestamp(time.Unix(1370818007, 0)),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createFakeServer(tt.handler)
			defer server.Close()
			httpClient := testingHTTPClient(server)

			cli := NewClient("", "", SetHTTPClient(httpClient))
			got, err := cli.Trades("btc_usd", 1)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Trades() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.Trades() = %v, want %v", got, tt.want)
			}
		})
	}
}
