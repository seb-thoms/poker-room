package game

import (
	"fmt"
	"math/rand"
)

// Suit represents a card suit
type Suit int

const (
	Clubs Suit = iota
	Diamonds
	Hearts
	Spades
)

// Rank represents a card rank
type Rank int

const (
	Two Rank = iota + 2
	Three
	Four
	Five
	Six
	Seven
	Eight
	Nine
	Ten
	Jack
	Queen
	King
	Ace
)

// Card represents a playing card
type Card struct {
	Suit Suit `json:"suit"`
	Rank Rank `json:"rank"`
}

// Deck represents a deck of cards
type Deck struct {
	cards []Card
	used  int
}

// NewDeck creates a standard 52-card deck
func NewDeck() *Deck {
	cards := make([]Card, 0, 52)

	for suit := Clubs; suit <= Spades; suit++ {
		for rank := Two; rank <= Ace; rank++ {
			cards = append(cards, Card{Suit: suit, Rank: rank})
		}
	}

	return &Deck{
		cards: cards,
		used:  0,
	}
}

// Shuffle randomizes the deck
func (d *Deck) Shuffle() {
	// Fisher-Yates shuffle
	for i := len(d.cards) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		d.cards[i], d.cards[j] = d.cards[j], d.cards[i]
	}
	d.used = 0
}

// Draw takes a card from the deck
func (d *Deck) Draw() Card {
	if d.used >= len(d.cards) {
		panic("no cards left in deck")
	}

	card := d.cards[d.used]
	d.used++
	return card
}

// CardsRemaining returns the number of cards left
func (d *Deck) CardsRemaining() int {
	return len(d.cards) - d.used
}

// Reset resets the deck without shuffling
func (d *Deck) Reset() {
	d.used = 0
}

// String representation methods

// String returns a string representation of the card
func (c Card) String() string {
	return fmt.Sprintf("%s%s", c.rankString(), c.suitString())
}

// ShortString returns a two-character representation (e.g., "As", "Kh")
func (c Card) ShortString() string {
	rank := ""
	switch c.Rank {
	case Ace:
		rank = "A"
	case King:
		rank = "K"
	case Queen:
		rank = "Q"
	case Jack:
		rank = "J"
	case Ten:
		rank = "T"
	default:
		rank = fmt.Sprintf("%d", c.Rank)
	}

	suit := ""
	switch c.Suit {
	case Spades:
		suit = "s"
	case Hearts:
		suit = "h"
	case Diamonds:
		suit = "d"
	case Clubs:
		suit = "c"
	}

	return rank + suit
}

func (c Card) rankString() string {
	switch c.Rank {
	case Ace:
		return "Ace"
	case King:
		return "King"
	case Queen:
		return "Queen"
	case Jack:
		return "Jack"
	case Ten:
		return "Ten"
	case Nine:
		return "Nine"
	case Eight:
		return "Eight"
	case Seven:
		return "Seven"
	case Six:
		return "Six"
	case Five:
		return "Five"
	case Four:
		return "Four"
	case Three:
		return "Three"
	case Two:
		return "Two"
	default:
		return "Unknown"
	}
}

func (c Card) suitString() string {
	switch c.Suit {
	case Spades:
		return "♠"
	case Hearts:
		return "♥"
	case Diamonds:
		return "♦"
	case Clubs:
		return "♣"
	default:
		return "?"
	}
}

// String returns the string representation of a suit
func (s Suit) String() string {
	switch s {
	case Spades:
		return "spades"
	case Hearts:
		return "hearts"
	case Diamonds:
		return "diamonds"
	case Clubs:
		return "clubs"
	default:
		return "unknown"
	}
}

// String returns the string representation of a rank
func (r Rank) String() string {
	switch r {
	case Ace:
		return "ace"
	case King:
		return "king"
	case Queen:
		return "queen"
	case Jack:
		return "jack"
	case Ten:
		return "ten"
	case Nine:
		return "nine"
	case Eight:
		return "eight"
	case Seven:
		return "seven"
	case Six:
		return "six"
	case Five:
		return "five"
	case Four:
		return "four"
	case Three:
		return "three"
	case Two:
		return "two"
	default:
		return "unknown"
	}
}

// MarshalJSON customizes JSON encoding for Card
func (c Card) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"rank":"%s","suit":"%s","display":"%s"}`,
		c.Rank.String(), c.Suit.String(), c.ShortString())), nil
}

// Utility functions for card comparisons

// CompareRank compares two cards by rank only
func CompareRank(a, b Card) int {
	if a.Rank < b.Rank {
		return -1
	} else if a.Rank > b.Rank {
		return 1
	}
	return 0
}

// CompareSuit compares two cards by suit only
func CompareSuit(a, b Card) int {
	if a.Suit < b.Suit {
		return -1
	} else if a.Suit > b.Suit {
		return 1
	}
	return 0
}

// SortCardsByRank sorts cards by rank (ascending)
func SortCardsByRank(cards []Card) {
	// Simple bubble sort for small arrays
	n := len(cards)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if cards[j].Rank > cards[j+1].Rank {
				cards[j], cards[j+1] = cards[j+1], cards[j]
			}
		}
	}
}

// SortCardsByRankDesc sorts cards by rank (descending)
func SortCardsByRankDesc(cards []Card) {
	n := len(cards)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if cards[j].Rank < cards[j+1].Rank {
				cards[j], cards[j+1] = cards[j+1], cards[j]
			}
		}
	}
}
