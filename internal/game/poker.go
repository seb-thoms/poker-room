package game

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

// PokerGame represents a poker game instance
type PokerGame struct {
	// Game configuration
	SmallBlind int
	BigBlind   int

	// Players
	Players      []*PokerPlayer
	DealerIndex  int
	CurrentIndex int

	// Current hand state
	Deck           *Deck
	CommunityCards []Card
	Pot            int
	SidePots       []SidePot
	CurrentBet     int
	MinRaise       int

	// Betting round tracking
	BettingRound     BettingRound
	LastAggressor    string // Player ID who last bet/raised
	NumActivePlayers int

	// Hand lifecycle
	HandNumber   int
	HandComplete bool
	Winners      []Winner
}

// PokerPlayer represents a player in the game
type PokerPlayer struct {
	ID             string
	Name           string
	Chips          int
	HoleCards      []Card
	CurrentBet     int
	TotalBetInHand int
	HasActed       bool
	IsFolded       bool
	IsAllIn        bool
	IsActive       bool // Still in the game (has chips)
	SeatPosition   int
}

// SidePot represents a side pot in the game
type SidePot struct {
	Amount          int
	EligiblePlayers []string // Player IDs eligible for this pot
}

// Winner represents a hand winner
type Winner struct {
	PlayerID    string
	Amount      int
	HandRank    HandRank
	BestHand    []Card
	Description string
}

// BettingRound represents the current betting round
type BettingRound int

const (
	PreFlop BettingRound = iota
	Flop
	Turn
	River
	Showdown
)

// ActionType represents possible player actions
type ActionType int

const (
	Check ActionType = iota
	Call
	Bet
	Raise
	Fold
	AllIn
)

// NewPokerGame creates a new poker game
func NewPokerGame(smallBlind, bigBlind int) *PokerGame {
	rand.Seed(time.Now().UnixNano())

	return &PokerGame{
		SmallBlind:  smallBlind,
		BigBlind:    bigBlind,
		Players:     make([]*PokerPlayer, 0),
		DealerIndex: 0,
		Deck:        NewDeck(),
	}
}

// AddPlayer adds a player to the game
func (g *PokerGame) AddPlayer(id, name string, chips int) error {
	if len(g.Players) >= 6 {
		return errors.New("maximum 6 players allowed")
	}

	player := &PokerPlayer{
		ID:           id,
		Name:         name,
		Chips:        chips,
		IsActive:     true,
		SeatPosition: len(g.Players),
	}

	g.Players = append(g.Players, player)
	return nil
}

// StartNewHand starts a new hand
func (g *PokerGame) StartNewHand() error {
	if len(g.Players) < 2 {
		return errors.New("need at least 2 players to start")
	}

	// Reset for new hand
	g.HandNumber++
	g.HandComplete = false
	g.Pot = 0
	g.SidePots = nil
	g.CurrentBet = 0
	g.MinRaise = g.BigBlind
	g.BettingRound = PreFlop
	g.CommunityCards = nil
	g.Winners = nil
	g.LastAggressor = ""

	// Reset deck
	g.Deck = NewDeck()
	g.Deck.Shuffle()

	// Reset players
	activePlayers := 0
	for _, p := range g.Players {
		if p.Chips > 0 {
			p.IsActive = true
			p.IsFolded = false
			p.IsAllIn = false
			p.HasActed = false
			p.CurrentBet = 0
			p.TotalBetInHand = 0
			p.HoleCards = nil
			activePlayers++
		} else {
			p.IsActive = false
		}
	}

	g.NumActivePlayers = activePlayers

	if activePlayers < 2 {
		return errors.New("not enough players with chips")
	}

	// Move dealer button
	g.moveDealerButton()

	// Post blinds
	g.postBlinds()

	// Deal hole cards
	g.dealHoleCards()

	// Set current player (left of big blind)
	g.CurrentIndex = g.getNextActivePlayer(g.getBigBlindIndex())

	return nil
}

// ProcessAction processes a player action
func (g *PokerGame) ProcessAction(playerID string, action ActionType, amount int) error {
	// Validate it's the player's turn
	currentPlayer := g.Players[g.CurrentIndex]
	if currentPlayer.ID != playerID {
		return errors.New("not your turn")
	}

	if currentPlayer.IsFolded || currentPlayer.IsAllIn {
		return errors.New("player cannot act")
	}

	// Process the action
	switch action {
	case Check:
		if g.CurrentBet > currentPlayer.CurrentBet {
			return errors.New("cannot check, must call or fold")
		}

	case Call:
		callAmount := g.CurrentBet - currentPlayer.CurrentBet
		if callAmount <= 0 {
			return errors.New("nothing to call")
		}
		g.playerBet(currentPlayer, callAmount)

	case Bet:
		if g.CurrentBet > 0 {
			return errors.New("cannot bet, must raise")
		}
		if amount < g.BigBlind {
			return errors.New("bet must be at least big blind")
		}
		g.playerBet(currentPlayer, amount)
		g.CurrentBet = amount
		g.MinRaise = amount
		g.LastAggressor = playerID
		g.resetHasActed()

	case Raise:
		if g.CurrentBet == 0 {
			return errors.New("cannot raise, must bet")
		}
		raiseAmount := amount - currentPlayer.CurrentBet
		if raiseAmount < g.MinRaise {
			return fmt.Errorf("raise must be at least %d", g.MinRaise)
		}
		g.playerBet(currentPlayer, amount-currentPlayer.CurrentBet)
		g.MinRaise = amount - g.CurrentBet
		g.CurrentBet = amount
		g.LastAggressor = playerID
		g.resetHasActed()

	case Fold:
		currentPlayer.IsFolded = true
		g.NumActivePlayers--

	case AllIn:
		allInAmount := currentPlayer.Chips
		g.playerBet(currentPlayer, allInAmount)
		if currentPlayer.CurrentBet > g.CurrentBet {
			g.MinRaise = currentPlayer.CurrentBet - g.CurrentBet
			g.CurrentBet = currentPlayer.CurrentBet
			g.LastAggressor = playerID
			g.resetHasActed()
		}
	}

	currentPlayer.HasActed = true

	// Check if betting round is complete
	if g.isBettingRoundComplete() {
		g.endBettingRound()
	} else {
		// Move to next player
		g.CurrentIndex = g.getNextActivePlayer(g.CurrentIndex)
	}

	return nil
}

// GetState returns the current game state
func (g *PokerGame) GetState() *GameState {
	players := make([]PlayerState, len(g.Players))
	for i, p := range g.Players {
		players[i] = PlayerState{
			ID:           p.ID,
			Name:         p.Name,
			Chips:        p.Chips,
			CurrentBet:   p.CurrentBet,
			IsFolded:     p.IsFolded,
			IsAllIn:      p.IsAllIn,
			IsActive:     p.IsActive,
			SeatPosition: p.SeatPosition,
			HasActed:     p.HasActed,
		}
	}

	return &GameState{
		Players:         players,
		CurrentPlayerID: g.Players[g.CurrentIndex].ID,
		DealerIndex:     g.DealerIndex,
		Pot:             g.Pot,
		CurrentBet:      g.CurrentBet,
		MinRaise:        g.MinRaise,
		CommunityCards:  g.CommunityCards,
		BettingRound:    g.getBettingRoundString(),
		HandComplete:    g.HandComplete,
		Winners:         g.Winners,
		HandNumber:      g.HandNumber,
		SidePots:        g.SidePots,
	}
}

// GetPlayerCards returns a player's hole cards (only for that player)
func (g *PokerGame) GetPlayerCards(playerID string) []Card {
	for _, p := range g.Players {
		if p.ID == playerID && !p.IsFolded {
			return p.HoleCards
		}
	}
	return nil
}

// Private helper methods

func (g *PokerGame) moveDealerButton() {
	// Find next active player to be dealer
	g.DealerIndex = g.getNextActivePlayer(g.DealerIndex)
}

func (g *PokerGame) getSmallBlindIndex() int {
	return g.getNextActivePlayer(g.DealerIndex)
}

func (g *PokerGame) getBigBlindIndex() int {
	return g.getNextActivePlayer(g.getSmallBlindIndex())
}

func (g *PokerGame) getNextActivePlayer(from int) int {
	next := (from + 1) % len(g.Players)
	for next != from {
		if g.Players[next].IsActive && !g.Players[next].IsFolded && !g.Players[next].IsAllIn {
			return next
		}
		next = (next + 1) % len(g.Players)
	}
	return from
}

func (g *PokerGame) postBlinds() {
	// Small blind
	sbIndex := g.getSmallBlindIndex()
	sbPlayer := g.Players[sbIndex]
	sbAmount := g.SmallBlind
	if sbAmount > sbPlayer.Chips {
		sbAmount = sbPlayer.Chips
	}
	g.playerBet(sbPlayer, sbAmount)

	// Big blind
	bbIndex := g.getBigBlindIndex()
	bbPlayer := g.Players[bbIndex]
	bbAmount := g.BigBlind
	if bbAmount > bbPlayer.Chips {
		bbAmount = bbPlayer.Chips
	}
	g.playerBet(bbPlayer, bbAmount)

	g.CurrentBet = g.BigBlind
}

func (g *PokerGame) dealHoleCards() {
	// Deal 2 cards to each active player
	for i := 0; i < 2; i++ {
		for _, p := range g.Players {
			if p.IsActive {
				card := g.Deck.Draw()
				p.HoleCards = append(p.HoleCards, card)
			}
		}
	}
}

func (g *PokerGame) playerBet(player *PokerPlayer, amount int) {
	if amount >= player.Chips {
		// All in
		amount = player.Chips
		player.IsAllIn = true
	}

	player.Chips -= amount
	player.CurrentBet += amount
	player.TotalBetInHand += amount
	g.Pot += amount
}

func (g *PokerGame) resetHasActed() {
	for _, p := range g.Players {
		if !p.IsFolded && !p.IsAllIn {
			p.HasActed = false
		}
	}
}

func (g *PokerGame) isBettingRoundComplete() bool {
	// Check if only one player left
	if g.NumActivePlayers == 1 {
		return true
	}

	// Check if all active players have acted and bets are equal
	for _, p := range g.Players {
		if p.IsActive && !p.IsFolded && !p.IsAllIn {
			if !p.HasActed {
				return false
			}
			if p.CurrentBet < g.CurrentBet {
				return false
			}
		}
	}

	return true
}

func (g *PokerGame) endBettingRound() {
	// Reset for next round
	for _, p := range g.Players {
		p.CurrentBet = 0
		p.HasActed = false
	}
	g.CurrentBet = 0

	// Create side pots if needed
	g.createSidePots()

	// Check if hand should end
	if g.shouldEndHand() {
		g.endHand()
		return
	}

	// Deal community cards
	switch g.BettingRound {
	case PreFlop:
		// Deal flop (3 cards)
		for i := 0; i < 3; i++ {
			g.CommunityCards = append(g.CommunityCards, g.Deck.Draw())
		}
		g.BettingRound = Flop

	case Flop:
		// Deal turn (1 card)
		g.CommunityCards = append(g.CommunityCards, g.Deck.Draw())
		g.BettingRound = Turn

	case Turn:
		// Deal river (1 card)
		g.CommunityCards = append(g.CommunityCards, g.Deck.Draw())
		g.BettingRound = River

	case River:
		// Go to showdown
		g.endHand()
		return
	}

	// Set next player
	g.CurrentIndex = g.getNextActivePlayer(g.DealerIndex)
}

func (g *PokerGame) shouldEndHand() bool {
	// Count non-folded players
	activePlayers := 0
	for _, p := range g.Players {
		if p.IsActive && !p.IsFolded {
			activePlayers++
		}
	}

	return activePlayers <= 1
}

func (g *PokerGame) createSidePots() {
	// Implementation for side pots when players are all-in
	// This is complex - simplified version for now
}

func (g *PokerGame) endHand() {
	g.HandComplete = true
	g.determineWinners()
	g.awardPots()
}

func (g *PokerGame) determineWinners() {
	// Find all non-folded players
	var contenders []*PokerPlayer
	for _, p := range g.Players {
		if p.IsActive && !p.IsFolded {
			contenders = append(contenders, p)
		}
	}

	if len(contenders) == 1 {
		// Only one player left, they win
		winner := contenders[0]
		g.Winners = []Winner{{
			PlayerID:    winner.ID,
			Amount:      g.Pot,
			Description: "Last player standing",
		}}
		return
	}

	// Evaluate hands at showdown
	type PlayerHand struct {
		Player   *PokerPlayer
		BestHand HandResult
	}

	var playerHands []PlayerHand
	for _, p := range contenders {
		allCards := append(p.HoleCards, g.CommunityCards...)
		bestHand := EvaluateBestHand(allCards)
		playerHands = append(playerHands, PlayerHand{
			Player:   p,
			BestHand: bestHand,
		})
	}

	// Sort by hand rank (best first)
	// In a real implementation, we'd sort properly
	// For now, find the best hand
	var bestRank HandRank = HighCard
	for _, ph := range playerHands {
		if ph.BestHand.Rank > bestRank {
			bestRank = ph.BestHand.Rank
		}
	}

	// Find all players with the best rank
	var winners []PlayerHand
	for _, ph := range playerHands {
		if ph.BestHand.Rank == bestRank {
			winners = append(winners, ph)
		}
	}

	// Split pot among winners
	potShare := g.Pot / len(winners)
	for _, w := range winners {
		g.Winners = append(g.Winners, Winner{
			PlayerID:    w.Player.ID,
			Amount:      potShare,
			HandRank:    w.BestHand.Rank,
			BestHand:    w.BestHand.Cards,
			Description: w.BestHand.Description,
		})
	}
}

func (g *PokerGame) awardPots() {
	// Award pots to winners
	for _, winner := range g.Winners {
		for _, p := range g.Players {
			if p.ID == winner.PlayerID {
				p.Chips += winner.Amount
				break
			}
		}
	}
}

func (g *PokerGame) getBettingRoundString() string {
	switch g.BettingRound {
	case PreFlop:
		return "preflop"
	case Flop:
		return "flop"
	case Turn:
		return "turn"
	case River:
		return "river"
	case Showdown:
		return "showdown"
	default:
		return "unknown"
	}
}

// ParseActionType converts string to ActionType
func ParseActionType(action string) (ActionType, error) {
	switch action {
	case "check":
		return Check, nil
	case "call":
		return Call, nil
	case "bet":
		return Bet, nil
	case "raise":
		return Raise, nil
	case "fold":
		return Fold, nil
	case "allin":
		return AllIn, nil
	default:
		return -1, fmt.Errorf("unknown action: %s", action)
	}
}

// GameState represents the public game state
type GameState struct {
	Players         []PlayerState `json:"players"`
	CurrentPlayerID string        `json:"currentPlayerId"`
	DealerIndex     int           `json:"dealerIndex"`
	Pot             int           `json:"pot"`
	CurrentBet      int           `json:"currentBet"`
	MinRaise        int           `json:"minRaise"`
	CommunityCards  []Card        `json:"communityCards"`
	BettingRound    string        `json:"bettingRound"`
	HandComplete    bool          `json:"handComplete"`
	Winners         []Winner      `json:"winners,omitempty"`
	HandNumber      int           `json:"handNumber"`
	SidePots        []SidePot     `json:"sidePots,omitempty"`
}

// PlayerState represents public player state
type PlayerState struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Chips        int    `json:"chips"`
	CurrentBet   int    `json:"currentBet"`
	IsFolded     bool   `json:"isFolded"`
	IsAllIn      bool   `json:"isAllIn"`
	IsActive     bool   `json:"isActive"`
	SeatPosition int    `json:"seatPosition"`
	HasActed     bool   `json:"hasActed"`
}
