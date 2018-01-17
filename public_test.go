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
	type args struct {
		result interface{}
		method string
	}
	tests := []struct {
		name    string
		args    args
		handler http.Handler
		wantErr bool
		errText string
	}{
		{
			name: "error response",
			args: args{
				result: &baseResponse{},
				method: "invalid",
			},
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, invalidMethodResponse)
			}),
			wantErr: true,
			errText: "server respond with error: Invalid method",
		},
		{
			name: "error code",
			args: args{
				result: &baseResponse{},
				method: "invalid",
			},
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}),
			wantErr: true,
			errText: "server respond with status code 500",
		},
		{
			name: "invalid json",
			args: args{
				result: &baseResponse{},
				method: "invalid",
			},
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, "/")
			}),
			wantErr: true,
			errText: "unmarshal to base response: invalid character '/' looking for beginning of value",
		},
		{
			name: "valid json",
			args: args{
				result: &baseResponse{},
				method: "invalid",
			},
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

			err := cli.publicRequest(tt.args.result, tt.args.method, nil)

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
						MinPrice:      decimal.NewFromFloat(0.1),
						MaxPrice:      decimal.NewFromFloat(400),
						MinAmount:     decimal.NewFromFloat(0.01),
						Fee:           decimal.NewFromFloat(0.2),
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
				High:             decimal.NewFromFloat(109.88),
				Low:              decimal.NewFromFloat(91.14),
				Average:          decimal.NewFromFloat(100.51),
				Volume:           decimal.NewFromFloat(1632898.2249),
				VolumeInCurrency: decimal.NewFromFloat(16541.51969),
				Last:             decimal.NewFromFloat(101.773),
				Buy:              decimal.NewFromFloat(101.9),
				Sell:             decimal.NewFromFloat(101.773),
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
						Rate:   decimal.NewFromFloat(103.426),
						Amount: decimal.NewFromFloat(0.01),
						Total:  decimal.NewFromFloat(1.03426),
					},
				},
				Bids: []Order{
					Order{
						Rate:   decimal.NewFromFloat(103.2),
						Amount: decimal.NewFromFloat(2.48502251),
						Total:  decimal.NewFromFloat(256.454323032),
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
					Rate:      decimal.NewFromFloat(103.6),
					Amount:    decimal.NewFromFloat(0.101),
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
