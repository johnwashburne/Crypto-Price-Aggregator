package exchange

import (
	"fmt"
	"log"
	"strconv"
)

type Exchange interface {
	Recv() error
	Name() string
}

type MarketUpdate struct {
	Bid       string
	Ask       string
	BidVolume string
	AskVolume string
}

func stringMultiply(s1 string, s2 string) string {
	f1, err := strconv.ParseFloat(s1, 64)
	if err != nil {
		log.Panicln("could not convert string to float")
	}

	f2, err := strconv.ParseFloat(s2, 64)
	if err != nil {
		log.Panicln("could not convert string to float")
	}

	return fmt.Sprintf("%f", f1*f2)
}
