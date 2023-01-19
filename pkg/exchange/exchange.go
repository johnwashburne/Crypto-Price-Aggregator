package exchange

import (
	"fmt"
	"log"
	"strconv"
)

const updateBufSize = 100

type Exchange interface {
	Recv()
	Updates() chan MarketUpdate
	Valid() bool
	Name() string
}

type MarketUpdate struct {
	Bid     string
	Ask     string
	BidSize string
	AskSize string
	Name    string
}

func stringMultiply(s1 string, s2 string) string {
	// need to create more efficient process
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
