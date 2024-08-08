package cards

import (
	b64 "encoding/base64"
	"fmt"
	"strings"

	"github.com/heroiclabs/nakama-common/runtime"
)

const DECK_SIZE = 32
const CARD_LIMIT = 4

type DeckDecodeError struct {
	Err     error
	Message string
}

func (e *DeckDecodeError) Error() string {
	return fmt.Sprintf("%s - %s", e.Err, e.Message)
}

func DecodeDeckCode(deck_code string, logger runtime.Logger) ([]string, DeckDecodeError) {

	cards := make([]string, DECK_SIZE)[:0] // Set capacity=DECK_SIZE and length=0

	sDec, err := b64.StdEncoding.DecodeString(deck_code)
	if err != nil {
		return cards, DeckDecodeError{
			err,
			"Failed b64 decode",
		}
	}

	buckets := strings.Split(string(sDec), "@")
	for card_amt, bucket := range buckets {
		bucket_cards := strings.Split(bucket, ":")
		if len(bucket_cards) == 0 {
			continue
		}
		for _, card_string := range bucket_cards {
			if card_string == "" {
				continue
			}
			logger.Warn(card_string)
			for n := range CARD_LIMIT {
				cards = append(cards, card_string)
				if n+1 >= card_amt || len(cards) >= DECK_SIZE {
					break
				}
			}
		}
	}
	logger.Warn("%s", cards)
	return cards, DeckDecodeError{nil, ""}
}
