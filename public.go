package wexapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

const (
	publicAPIEndpoint = "https://wex.nz/api/3"

	orderParamsCount = 2
	orderRateIndex   = 0
	orderAmountIndex = 1
)

// InfoResponse for /info path.
type InfoResponse struct {
	ServerTime unixTimestamp       `json:"server_time"`
	Pairs      map[string]PairInfo `json:"pairs"`
}

// PairInfo struct.
type PairInfo struct {
	DecimalPlaces uint32          `json:"decimal_places"`
	MinPrice      decimal.Decimal `json:"min_price"`
	MaxPrice      decimal.Decimal `json:"max_price"`
	MinAmount     decimal.Decimal `json:"min_amount"`
	Fee           decimal.Decimal `json:"fee"`
	Hidden        convertibleBool `json:"hidden"`
}

// Info provides all the information about currently
// active pairs, such as the maximum number of digits
// after the decimal point, the minimum price, the
// maximum price, the minimum transaction size, whether
// the pair is hidden, the commission for each pair.
func (cli *Client) Info() (InfoResponse, error) {
	infoResponse := InfoResponse{}
	err := cli.publicRequest(&infoResponse, "info", nil)
	return infoResponse, err
}

// Market holds data about pair.
type Market struct {
	High             decimal.Decimal `json:"high"`
	Low              decimal.Decimal `json:"low"`
	Average          decimal.Decimal `json:"avg"`
	Volume           decimal.Decimal `json:"vol"`
	VolumeInCurrency decimal.Decimal `json:"vol_cur"`
	Last             decimal.Decimal `json:"last"`
	Buy              decimal.Decimal `json:"buy"`
	Sell             decimal.Decimal `json:"sell"`
	Updated          unixTimestamp   `json:"updated"`
}

// Ticker provides all the information about currently
// active pairs, such as: the maximum price, the minimum
// price, average price, trade volume, trade volume in
// currency, the last trade, Buy and Sell price.
func (cli *Client) Ticker(pair string) (Market, error) {
	tickerResponse := make(map[string]Market)
	err := cli.publicRequest(&tickerResponse, fmt.Sprintf("ticker/%s", pair), nil)
	return tickerResponse[pair], err
}

// Order holds data about order.
type Order struct {
	Rate   decimal.Decimal
	Amount decimal.Decimal
	Total  decimal.Decimal
}

// UnmarshalJSON parses [103.426,0.01] format into Order.
func (order *Order) UnmarshalJSON(data []byte) error {
	var orderParams [orderParamsCount]decimal.Decimal

	if err := json.Unmarshal(data, &orderParams); err != nil {
		return err
	}

	order.Rate = orderParams[orderRateIndex]
	order.Amount = orderParams[orderAmountIndex]
	order.CalculateTotal()

	return nil
}

// CalculateTotal calculates total amount
// by multiplying rate and amount.
func (order *Order) CalculateTotal() {
	if order.Amount.Equal(decimal.Zero) {
		return
	}

	order.Total = order.Rate.Mul(order.Amount)
}

// OrderBook holds data about pair orders.
type OrderBook struct {
	Asks []Order `json:"asks"`
	Bids []Order `json:"bids"`
}

// Depth provides the information about active
// orders on the pair.
func (cli *Client) Depth(pair string, limit int) (OrderBook, error) {
	depthResponse := make(map[string]OrderBook)
	param := param{key: "limit", value: strconv.Itoa(limit)}
	err := cli.publicRequest(&depthResponse, fmt.Sprintf("depth/%s", pair), &param)
	return depthResponse[pair], err
}

// Trade holds data about trade.
type Trade struct {
	ID        uint64          `json:"tid"`
	Type      string          `json:"type"`
	Rate      decimal.Decimal `json:"price"`
	Amount    decimal.Decimal `json:"amount"`
	Timestamp unixTimestamp   `json:"timestamp"`
}

// Trades provides the information about the last trades.
func (cli *Client) Trades(pair string, limit int) ([]Trade, error) {
	tradeResponse := make(map[string][]Trade)
	param := param{key: "limit", value: strconv.Itoa(limit)}
	err := cli.publicRequest(&tradeResponse, fmt.Sprintf("trades/%s", pair), &param)
	return tradeResponse[pair], err
}

func (cli *Client) publicRequest(result interface{}, method string, prm *param) error {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", publicAPIEndpoint, method), nil)
	if err != nil {
		return errors.Wrap(err, "request build")
	}

	if prm != nil {
		q := url.Values{}
		q.Add(prm.key, prm.value)
		req.URL.RawQuery = q.Encode()
	}

	resp, err := cli.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "do request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("server respond with status code %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "read response body")
	}

	br := baseResponse{}
	if err := json.Unmarshal(body, &br); err != nil {
		return errors.Wrap(err, "unmarshal to base response")
	}

	if !br.Success && br.Error != nil {
		return errors.Errorf("server respond with error: %s", *br.Error)
	}

	err = json.Unmarshal(body, result)
	return errors.Wrap(err, "unmarshal to result")
}

type baseResponse struct {
	Success convertibleBool `json:"success"`
	Error   *string         `json:"error"`
	Return  json.RawMessage `json:"return"`
}

type param struct {
	key, value string
}
