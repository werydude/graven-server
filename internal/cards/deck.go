package cards

import (
	b64 "encoding/base64"
	"fmt"
	"strconv"
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

type DeckEncodeError struct {
	Err     error
	Message string
}

func (e *DeckEncodeError) Error() string {
	return fmt.Sprintf("%s - %s", e.Err, e.Message)
}

func DecodeDeckCode(deck_code string, instances_ptr *string, logger runtime.Logger) ([]InstanceCard, DeckDecodeError) {

	cards := make([]InstanceCard, DECK_SIZE)[:0] // Set capacity=DECK_SIZE and length=0

	sDec, err := b64.StdEncoding.DecodeString(deck_code)
	if err != nil {
		return cards, DeckDecodeError{
			err,
			"Failed b64 decode",
		}
	}

	var instances []string
	if instances_ptr != nil {
		instances = strings.Split(*instances_ptr, "|")
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
			//logger.Warn(card_string)
			for n := range CARD_LIMIT {
				instance_id := -1
				if instances_ptr != nil {
					i, _ := strconv.Atoi(instances[len(cards)])
					instance_id = i
				}
				cards = append(cards, InstanceCard{CardId(card_string), instance_id})
				if n+1 >= card_amt || len(cards) >= DECK_SIZE {
					break
				}
			}
		}
	}
	logger.Warn("%s", cards)
	return cards, DeckDecodeError{nil, ""}
}

func EncodeDeckCode(deck []InstanceCard, logger runtime.Logger) ([]byte, DeckEncodeError) {
	card_id_list := make(map[CardId]int)

	for _, card := range deck {
		_, err := card_id_list[card.CardId]
		if !err {
			card_id_list[card.CardId] = 1
		} else {
			card_id_list[card.CardId] += 1
		}
	}
	logger.Warn("%s", card_id_list)
	buckets := make([][]string, 5)
	for k := range card_id_list {
		if strings.ContainsRune(string(k), 'T') {
			buckets[0] = append(buckets[0], string(k))
		} else {
			buckets[card_id_list[k]] = append(buckets[card_id_list[k]], string(k))
		}
	}

	bucket_str := [5]string{
		strings.Join(buckets[0], ":"),
		strings.Join(buckets[1], ":"),
		strings.Join(buckets[2], ":"),
		strings.Join(buckets[3], ":"),
		strings.Join(buckets[4], ":"),
	}

	full_string := strings.Join(bucket_str[:], "@")
	return []byte(b64.StdEncoding.EncodeToString([]byte(full_string))), DeckEncodeError{nil, ""}
}
