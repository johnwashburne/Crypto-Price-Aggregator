package aggregator

import (
	"log"

	"github.com/johnwashburne/Crypto-Price-Aggregator/exchange"
)

type BestPrice struct {
	Bid         string
	BidSize     string
	BidPlatform string
	Ask         string
	AskSize     string
	AskPlatform string
}

type Aggregator struct {
	updates   chan BestPrice
	exchanges []exchange.Exchange
}

// Create a new aggregator struct
func New(exchanges ...exchange.Exchange) Aggregator {
	c := make(chan BestPrice, 100)
	return Aggregator{
		updates:   c,
		exchanges: exchanges,
	}
}

// receive and aggregate updates
// send BestPrice over the Updates channel when an update to the best bid or ask occurs
func (a *Aggregator) Recv() {

	// channel that receives MarketUpdates for all exchanges
	agg := make(chan exchange.MarketUpdate, 100)

	// track the current top of book for all exchanges
	topOfBook := make(map[string]exchange.MarketUpdate)

	price := BestPrice{}
	lastPrice := BestPrice{}

	for _, exch := range a.exchanges {
		if !exch.Valid() {
			log.Println(exch.Name(), "not valid, cannot connect")
			continue
		}

		go exch.Recv()
		go func(c chan exchange.MarketUpdate) {
			for msg := range c {
				agg <- msg
			}
		}(exch.Updates())
	}

	for msg := range agg {
		topOfBook[msg.Name] = msg

		if msg.Name == price.AskPlatform || msg.Name == price.BidPlatform {
			// if there is an update to the top of book for current best bid or best ask
			// must iterate through top of all exchanges in case there was a match
			price = BestPrice{}
			for _, data := range topOfBook {
				compare(&price, data)
			}
		} else {
			// else, simply compare the best bid and best ask with this most recent update
			compare(&price, msg)
		}

		if price != lastPrice {
			a.updates <- price
			lastPrice = price
		}
	}
}

func (a *Aggregator) Updates() chan BestPrice {
	return a.updates
}

func compare(price *BestPrice, update exchange.MarketUpdate) {
	if price.Bid == "" || price.Bid < update.Bid || (price.Bid == update.Bid && price.BidSize < update.BidSize) {
		price.Bid = update.Bid
		price.BidSize = update.BidSize
		price.BidPlatform = update.Name
	}

	if update.Ask != "" && (price.Ask == "" || price.Ask > update.Ask || (price.Ask == update.Ask && price.AskSize < update.AskSize)) {
		price.Ask = update.Ask
		price.AskSize = update.AskSize
		price.AskPlatform = update.Name
	}
}
