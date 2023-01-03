package exchange

import (
	"fmt"
	"log"
	"strconv"
)

const updateBufSize = 100

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
