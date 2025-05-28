package game

import (
	"fmt"
)

// HandRank represents the rank of a poker hand
type HandRank int

const (
	HighCard HandRank = iota
	OnePair
	TwoPair
	ThreeOfAKind
	Straight
	Flush
	FullHouse
	FourOfAKind
	StraightFlush
	RoyalFlush
)

// HandResult represents the result of hand evaluation
type HandResult struct {
	Rank        HandRank
	Cards       []Card // The 5 cards that make up the hand
	Kickers     []Card // Kicker cards for tie breaking
	Description string // Human-readable description
}

// EvaluateBestHand finds the best 5-card hand from available cards
func EvaluateBestHand(cards []Card) HandResult {
	if len(cards) < 5 {
		return HandResult{Rank: HighCard, Description: "Not enough cards"}
	}

	// Generate all possible 5-card combinations
	bestResult := HandResult{Rank: HighCard}

	// For Texas Hold'em with 7 cards (2 hole + 5 community)
	// we need to check C(7,5) = 21 combinations
	combinations := generateCombinations(cards, 5)

	for _, combo := range combinations {
		result := evaluateFiveCards(combo)
		if result.Rank > bestResult.Rank {
			bestResult = result
		} else if result.Rank == bestResult.Rank {
			// Need to compare kickers
			if compareHands(result, bestResult) > 0 {
				bestResult = result
			}
		}
	}

	return bestResult
}

// evaluateFiveCards evaluates exactly 5 cards
func evaluateFiveCards(cards []Card) HandResult {
	if len(cards) != 5 {
		panic("evaluateFiveCards requires exactly 5 cards")
	}

	// Make a copy to avoid modifying original
	hand := make([]Card, 5)
	copy(hand, cards)

	// Sort by rank (descending)
	SortCardsByRankDesc(hand)

	// Check for flush
	isFlush := true
	for i := 1; i < 5; i++ {
		if hand[i].Suit != hand[0].Suit {
			isFlush = false
			break
		}
	}

	// Check for straight
	isStraight := checkStraight(hand)

	// Special case: Ace-low straight (A-2-3-4-5)
	isWheelStraight := false
	if !isStraight && hand[0].Rank == Ace {
		// Check for A-5-4-3-2
		if hand[1].Rank == Five && hand[2].Rank == Four &&
			hand[3].Rank == Three && hand[4].Rank == Two {
			isWheelStraight = true
			// Rearrange to 5-4-3-2-A for proper ordering
			hand = []Card{hand[1], hand[2], hand[3], hand[4], hand[0]}
		}
	}

	isStraight = isStraight || isWheelStraight

	// Count ranks
	rankCounts := make(map[Rank]int)
	for _, card := range hand {
		rankCounts[card.Rank]++
	}

	// Find pairs, trips, quads
	var pairs []Rank
	var trips []Rank
	var quads []Rank

	for rank, count := range rankCounts {
		switch count {
		case 2:
			pairs = append(pairs, rank)
		case 3:
			trips = append(trips, rank)
		case 4:
			quads = append(quads, rank)
		}
	}

	// Determine hand rank
	if isFlush && isStraight {
		if hand[0].Rank == Ace && hand[1].Rank == King {
			return HandResult{
				Rank:        RoyalFlush,
				Cards:       hand,
				Description: fmt.Sprintf("Royal Flush in %s", hand[0].Suit),
			}
		}
		return HandResult{
			Rank:        StraightFlush,
			Cards:       hand,
			Description: fmt.Sprintf("Straight Flush, %s high", hand[0].rankString()),
		}
	}

	if len(quads) > 0 {
		return HandResult{
			Rank:        FourOfAKind,
			Cards:       hand,
			Description: fmt.Sprintf("Four of a Kind, %ss", Rank(quads[0]).String()),
		}
	}

	if len(trips) > 0 && len(pairs) > 0 {
		return HandResult{
			Rank:  FullHouse,
			Cards: hand,
			Description: fmt.Sprintf("Full House, %ss full of %ss",
				Rank(trips[0]).String(), Rank(pairs[0]).String()),
		}
	}

	if isFlush {
		return HandResult{
			Rank:        Flush,
			Cards:       hand,
			Description: fmt.Sprintf("Flush, %s high", hand[0].rankString()),
		}
	}

	if isStraight {
		return HandResult{
			Rank:        Straight,
			Cards:       hand,
			Description: fmt.Sprintf("Straight, %s high", hand[0].rankString()),
		}
	}

	if len(trips) > 0 {
		return HandResult{
			Rank:        ThreeOfAKind,
			Cards:       hand,
			Description: fmt.Sprintf("Three of a Kind, %ss", Rank(trips[0]).String()),
		}
	}

	if len(pairs) >= 2 {
		// Sort pairs by rank
		if len(pairs) > 1 && pairs[0] < pairs[1] {
			pairs[0], pairs[1] = pairs[1], pairs[0]
		}
		return HandResult{
			Rank:  TwoPair,
			Cards: hand,
			Description: fmt.Sprintf("Two Pair, %ss and %ss",
				Rank(pairs[0]).String(), Rank(pairs[1]).String()),
		}
	}

	if len(pairs) == 1 {
		return HandResult{
			Rank:        OnePair,
			Cards:       hand,
			Description: fmt.Sprintf("One Pair, %ss", Rank(pairs[0]).String()),
		}
	}

	return HandResult{
		Rank:        HighCard,
		Cards:       hand,
		Description: fmt.Sprintf("High Card, %s", hand[0].rankString()),
	}
}

// checkStraight checks if 5 cards form a straight
func checkStraight(cards []Card) bool {
	// Assumes cards are sorted by rank descending
	for i := 0; i < 4; i++ {
		if cards[i].Rank != cards[i+1].Rank+1 {
			return false
		}
	}
	return true
}

// compareHands compares two hands of the same rank
func compareHands(a, b HandResult) int {
	// For now, simplified comparison
	// In a full implementation, we'd compare kickers properly
	for i := 0; i < len(a.Cards) && i < len(b.Cards); i++ {
		if a.Cards[i].Rank > b.Cards[i].Rank {
			return 1
		} else if a.Cards[i].Rank < b.Cards[i].Rank {
			return -1
		}
	}
	return 0
}

// generateCombinations generates all combinations of k cards from n cards
func generateCombinations(cards []Card, k int) [][]Card {
	var result [][]Card
	n := len(cards)

	// Helper function for recursive generation
	var generate func(start int, current []Card)
	generate = func(start int, current []Card) {
		if len(current) == k {
			combo := make([]Card, k)
			copy(combo, current)
			result = append(result, combo)
			return
		}

		for i := start; i < n; i++ {
			generate(i+1, append(current, cards[i]))
		}
	}

	generate(0, []Card{})
	return result
}

// String returns string representation of hand rank
func (h HandRank) String() string {
	switch h {
	case RoyalFlush:
		return "Royal Flush"
	case StraightFlush:
		return "Straight Flush"
	case FourOfAKind:
		return "Four of a Kind"
	case FullHouse:
		return "Full House"
	case Flush:
		return "Flush"
	case Straight:
		return "Straight"
	case ThreeOfAKind:
		return "Three of a Kind"
	case TwoPair:
		return "Two Pair"
	case OnePair:
		return "One Pair"
	case HighCard:
		return "High Card"
	default:
		return "Unknown"
	}
}
