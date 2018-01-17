package main

import (
	"fmt"
	"log"

	"github.com/romanyx/wexapi"
)

func main() {
	cli := wexapi.NewClient("", "")
	info, err := cli.Info()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", info)

	market, err := cli.Ticker("eth_btc")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", market)

	depth, err := cli.Depth("eth_btc", 2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", depth)

	trades, err := cli.Trades("eth_btc", 2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", trades)
}
