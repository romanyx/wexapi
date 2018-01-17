package wexapi

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

const (
	tradeAPIEndpoint = "https://wex.nz/tapi"
)

// Rights is a privileges of the current API key.
type Rights struct {
	Info     uint64 `json:"info"`
	Trade    uint64 `json:"trade"`
	Withdraw uint64 `json:"withdraw"`
}

// Funds is a account balance available for trading
type Funds map[string]decimal.Decimal

// UserInfo is an information about the user’s current balance.
type UserInfo struct {
	Funds            Funds         `json:"funds"`
	Rights           Rights        `json:"rights"`
	TransactionCount uint64        `json:"transaction_count"`
	OpenOrders       uint64        `json:"open_orders"`
	ServerTime       unixTimestamp `json:"server_time"`
}

// GetInfo eturns information about the user’s current
// balance, API-key privileges, the number of open orders
// and Server Time.
// To use this method you need a privilege of the key info.
func (cli *Client) GetInfo() (UserInfo, error) {
	userInfo := UserInfo{}
	err := cli.tradeRequest(&userInfo, "getInfo")
	return userInfo, err
}

// UserTrade holds data about trade.
type UserTrade struct {
	Received decimal.Decimal `json:"received"`
	Remains  decimal.Decimal `json:"remains"`
	OrderID  uint64          `json:"order_id"`
	Funds    Funds           `json:"funds"`
}

// Trade is the basic method that can be used for
// creating orders and trading on the exchange.
// To use this method you need a privilege of the key info.
func (cli *Client) Trade(pair, tradeType string, rate, amount decimal.Decimal) (UserTrade, error) {
	userTrade := UserTrade{}
	params := []param{
		param{key: "pair", value: pair},
		param{key: "type", value: tradeType},
		param{key: "rate", value: rate.String()},
		param{key: "amount", value: amount.String()},
	}
	err := cli.tradeRequest(&userTrade, "Trade", params...)
	return userTrade, err
}

// TradeOrder holds information about user trade orders.
type TradeOrder struct {
	ID               uint64
	Pair             string          `json:"pair"`
	Type             string          `json:"type"`
	StartAmount      decimal.Decimal `json:"start_amount"`
	Amount           decimal.Decimal `json:"amount"`
	Rate             decimal.Decimal `json:"rate"`
	TimestampCreated unixTimestamp   `json:"timestamp_created"`
}

// TradeOrders holds list of trade orders.
type TradeOrders []TradeOrder

// UnmarshalJSON unmarshall map[string]TradeOrder
// format into the slice.
func (to *TradeOrders) UnmarshalJSON(data []byte) error {
	idRaw := make(map[string]json.RawMessage)

	if err := json.Unmarshal(data, &idRaw); err != nil {
		return errors.Wrap(err, "unmarshal to id raw")
	}

	for idString, data := range idRaw {
		id, err := strconv.ParseUint(idString, 10, 64)
		if err != nil {
			return errors.Wrapf(err, "parse id %s", idString)
		}

		trade := TradeOrder{
			ID: id,
		}
		if err := json.Unmarshal(data, &trade); err != nil {
			return errors.Wrapf(err, "unmarshal trade order %s", idString)
		}

		*to = append(*to, trade)
	}

	return nil
}

// ActiveOrders returns the list of your active orders.
// To use this method you need a privilege of the info key.
func (cli *Client) ActiveOrders(pair string) (TradeOrders, error) {
	tradeOrders := TradeOrders{}
	err := cli.tradeRequest(&tradeOrders, "ActiveOrders", param{key: "pair", value: pair})
	return tradeOrders, err
}

// OrderInfo returns the information on particular order.
// To use this method you need a privilege of the info key.
func (cli *Client) OrderInfo(orderID uint64) (TradeOrder, error) {
	tradeOrders := make(map[string]TradeOrder)
	orderIDString := strconv.FormatUint(orderID, 10)
	err := cli.tradeRequest(&tradeOrders, "OrderInfo", param{key: "order_id", value: orderIDString})
	trade := tradeOrders[orderIDString]
	trade.ID = orderID
	return tradeOrders[orderIDString], err
}

// CancelOrder holds data about cancelled order.
type CancelOrder struct {
	OrderID uint64 `json:"order_id"`
}

// CancelOrder returns the information on particular order.
// To use this method you need a privilege of the info key.
func (cli *Client) CancelOrder(orderID uint64) (CancelOrder, error) {
	cancelOrder := CancelOrder{}
	err := cli.tradeRequest(&cancelOrder, "CancelOrder", param{key: "order_id", value: strconv.FormatUint(orderID, 10)})
	return cancelOrder, err
}

// Withdraw holds data about withdraw.
type Withdraw struct {
	TradeID    uint64          `json:"tId"`
	AmountSent decimal.Decimal `json:"amountSent"`
	Funds      Funds           `json:"funds"`
}

// WithdrawCoin is designed for cryptocurrency withdrawals.
// To use this method you need a privilege of the info key.
func (cli *Client) WithdrawCoin(currency, address string, amount decimal.Decimal) (Withdraw, error) {
	withdraw := Withdraw{}
	params := []param{
		param{key: "coinName", value: currency},
		param{key: "address", value: address},
		param{key: "amount", value: amount.String()},
	}
	err := cli.tradeRequest(&withdraw, "WithdrawCoin", params...)
	return withdraw, err
}

func (cli *Client) tradeRequest(result interface{}, method string, params ...param) error {
	data := url.Values{
		"method": []string{method},
		"nonce":  []string{cli.nonce()},
	}

	for _, param := range params {
		data.Add(param.key, param.value)
	}

	buf := bytes.NewBufferString(data.Encode())
	req, err := http.NewRequest("POST", tradeAPIEndpoint, buf)
	if err != nil {
		return errors.Wrap(err, "request build")
	}

	sign := hmac.New(sha512.New, []byte(cli.secret))
	if _, err := sign.Write(buf.Bytes()); err != nil {
		return errors.Wrap(err, "hmac write signature")
	}

	req.Header.Set("Key", cli.key)
	req.Header.Set("Sign", hex.EncodeToString(sign.Sum(nil)))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

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

	err = json.Unmarshal(br.Return, result)
	return errors.Wrap(err, "unmarshal to result")
}
