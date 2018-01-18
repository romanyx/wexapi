package wexapi

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

const (
	invalidReturnResponse = `{"success":1,"return":"//"}`
	getInfoResponse       = `{
		"success":1,
		"return":{
			"funds":{
				"usd":325,
				"btc":23.998,
				"ltc":0
			},
			"rights":{
				"info":1,
				"trade":0,
				"withdraw":0
			},
			"transaction_count":0,
			"open_orders":1,
			"server_time":1342123547
		}
	}`
	tradeResponse = `{
		"success":1,
		"return":{
			"received":0.1,
			"remains":0,
			"order_id":0,
			"funds":{
				"usd":325,
				"btc":2.498,
				"ltc":0
			}
		}
	}`
	activeOrdersResponse = `{
		"success":1,
		"return":{
			"343152":{
				"pair":"btc_usd",
				"type":"sell",
				"amount":12.345,
				"rate":485,
				"timestamp_created":1342448420,
				"status":0
			}
		}
	}`
	orderInfoResponse = `{
		"success":1,
		"return":{
			"343152":{
				"pair":"btc_usd",
				"type":"sell",
				"start_amount":13.345,
				"amount":12.345,
				"rate":485,
				"timestamp_created":1342448420,
				"status":0
			}
		}
	}`
	withdrawResponse = `{
		"success":1,
		"return":{
			"tId":37832629,
			"amountSent":0.009,
			"funds":{
				"usd":325,
				"btc":24.998,
				"ltc":0
			}
		}
	}`
)

func TestClient_tradeRequest(t *testing.T) {
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
			name: "invalid return json",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, "/")
			}),
			wantErr: true,
			errText: "unmarshal to base response: invalid character '/' looking for beginning of value",
		},
		{
			name: "valid json",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, getInfoResponse)
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

			err := cli.tradeRequest(&baseResponse{}, "any")

			if tt.wantErr {
				if err == nil {
					t.Errorf("Client.tradeRequest() error = %s, wantErr %t", err, tt.wantErr)
				}

				if tt.errText != err.Error() {
					t.Errorf("expected err text: %s, got: %s", tt.errText, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Client.tradeRequest() error = %s, wantErr %t", err, tt.wantErr)
			}
		})
	}
}

func (f Funds) String() string {
	keys := make([]string, 0, len(f))

	for key := range f {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return strings.Join(keys, "")
}

func TestClient_GetInfo(t *testing.T) {
	tests := []struct {
		name    string
		handler http.Handler
		want    UserInfo
		wantErr bool
	}{
		{
			name: "valid json",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, getInfoResponse)
			}),
			want: UserInfo{
				Funds: map[string]decimal.Decimal{
					"usd": decimal.NewFromFloat(325),
					"btc": decimal.NewFromFloat(23.998),
					"ltc": decimal.Zero,
				},
				Rights: Rights{
					Info:     1,
					Trade:    0,
					Withdraw: 0,
				},
				TransactionCount: 0,
				OpenOrders:       1,
				ServerTime:       unixTimestamp(time.Unix(1342123547, 0)),
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
			got, err := cli.GetInfo()
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !compareAsStrings(got, tt.want) {
				t.Errorf("Client.GetInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Trade(t *testing.T) {
	tests := []struct {
		name    string
		handler http.Handler
		want    UserTrade
		wantErr bool
	}{
		{
			name: "valid json",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, tradeResponse)
			}),
			want: UserTrade{
				Funds: map[string]decimal.Decimal{
					"usd": decimal.NewFromFloat(325),
					"btc": decimal.NewFromFloat(2.498),
					"ltc": decimal.Zero,
				},
				Received: decimal.NewFromFloat(0.1),
				Remains:  decimal.Zero,
				OrderID:  0,
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
			got, err := cli.Trade("eth_btc", "sell", decimal.Zero, decimal.Zero)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Trade() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !compareAsStrings(got, tt.want) {
				t.Errorf("Client.Trade() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_ActiveOrders(t *testing.T) {
	tests := []struct {
		name    string
		handler http.Handler
		want    TradeOrders
		wantErr bool
	}{
		{
			name: "valid json",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, activeOrdersResponse)
			}),
			want: TradeOrders{
				TradeOrder{
					ID:               343152,
					Pair:             "btc_usd",
					Type:             "sell",
					Amount:           decimal.NewFromFloat(12.345),
					Rate:             decimal.NewFromFloat(485),
					TimestampCreated: unixTimestamp(time.Unix(1342448420, 0)),
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
			got, err := cli.ActiveOrders("btc_usd")
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.ActiveOrders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !compareAsStrings(got, tt.want) {
				t.Errorf("Client.ActiveOrders() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_OrderInfo(t *testing.T) {
	tests := []struct {
		name    string
		handler http.Handler
		want    TradeOrder
		wantErr bool
	}{
		{
			name: "valid json",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, orderInfoResponse)
			}),
			want: TradeOrder{
				Pair:             "btc_usd",
				Type:             "sell",
				StartAmount:      decimal.NewFromFloat(13.345),
				Amount:           decimal.NewFromFloat(12.345),
				Rate:             decimal.NewFromFloat(485),
				TimestampCreated: unixTimestamp(time.Unix(1342448420, 0)),
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
			got, err := cli.OrderInfo(343152)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.OrderInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !compareAsStrings(got, tt.want) {
				t.Errorf("Client.OrderInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_WithdrawCoin(t *testing.T) {
	tests := []struct {
		name    string
		handler http.Handler
		want    Withdraw
		wantErr bool
	}{
		{
			name: "valid json",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, withdrawResponse)
			}),
			want: Withdraw{
				TradeID:    37832629,
				AmountSent: decimal.NewFromFloat(0.009),
				Funds: Funds{
					"usd": decimal.NewFromFloat(325),
					"btc": decimal.NewFromFloat(24.998),
					"ltc": decimal.NewFromFloat(0),
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
			got, err := cli.WithdrawCoin("btc", "address", decimal.Zero)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.WithdrawCoin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !compareAsStrings(got, tt.want) {
				t.Errorf("Client.WithdrawCoin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func BenchmarkClient_GetInfo(b *testing.B) {
	server := createFakeServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, getInfoResponse)
	}))
	defer server.Close()
	httpClient := testingHTTPClient(server)
	cli := NewClient("", "", SetHTTPClient(httpClient))

	b.ResetTimer()
	for i := 0; i < b.N; i = i + 1 {
		cli.GetInfo()
	}
}
