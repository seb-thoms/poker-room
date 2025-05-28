package game

// Game configuration constants
const (
	// Table limits
	MinPlayers = 2
	MaxPlayers = 6

	// Default blinds
	DefaultSmallBlind = 10
	DefaultBigBlind   = 20

	// Default buy-in
	DefaultBuyIn = 10000
	MinBuyIn     = 1000
	MaxBuyIn     = 100000

	// Timing (in seconds)
	TurnTimeout       = 30
	DisconnectTimeout = 300 // 5 minutes
	HandPauseTime     = 5   // Pause between hands

	// Betting limits
	MinBetMultiplier = 1 // Minimum bet is 1x big blind
	MinRaiseAmount   = 1 // Minimum raise is previous bet/raise size
)

// GameRules represents configurable game rules
type GameRules struct {
	// Blinds
	SmallBlind int
	BigBlind   int

	// Buy-in rules
	MinBuyIn int
	MaxBuyIn int

	// Table rules
	MaxPlayers int
	MinPlayers int

	// Timing rules
	TurnTimeout       int // Seconds per turn
	DisconnectTimeout int // Seconds before folding disconnected player

	// Allow features
	AllowRebuy      bool
	AllowSitOut     bool
	AllowRunItTwice bool
}

// DefaultRules returns standard No-Limit Texas Hold'em rules
func DefaultRules() GameRules {
	return GameRules{
		SmallBlind:        DefaultSmallBlind,
		BigBlind:          DefaultBigBlind,
		MinBuyIn:          DefaultBuyIn,
		MaxBuyIn:          DefaultBuyIn,
		MaxPlayers:        MaxPlayers,
		MinPlayers:        MinPlayers,
		TurnTimeout:       TurnTimeout,
		DisconnectTimeout: DisconnectTimeout,
		AllowRebuy:        true,
		AllowSitOut:       true,
		AllowRunItTwice:   false,
	}
}

// ValidateAction checks if an action is valid given the current game state
func ValidateAction(
	action ActionType,
	amount int,
	playerChips int,
	currentBet int,
	playerBet int,
	minRaise int,
	potSize int,
) error {
	switch action {
	case Check:
		if currentBet > playerBet {
			return ErrCannotCheck
		}

	case Call:
		callAmount := currentBet - playerBet
		if callAmount <= 0 {
			return ErrNothingToCall
		}
		if callAmount > playerChips {
			return ErrInsufficientChips
		}

	case Bet:
		if currentBet > 0 {
			return ErrCannotBet
		}
		if amount < DefaultBigBlind {
			return ErrBetTooSmall
		}
		if amount > playerChips {
			return ErrInsufficientChips
		}

	case Raise:
		if currentBet == 0 {
			return ErrCannotRaise
		}
		totalAmount := amount
		raiseAmount := totalAmount - currentBet
		if raiseAmount < minRaise {
			return ErrRaiseTooSmall
		}
		if totalAmount-playerBet > playerChips {
			return ErrInsufficientChips
		}

	case Fold:
		// Always valid

	case AllIn:
		if playerChips <= 0 {
			return ErrNoChips
		}
	}

	return nil
}

// Common errors
var (
	ErrCannotCheck       = NewGameError("cannot check, must call or fold")
	ErrNothingToCall     = NewGameError("nothing to call")
	ErrCannotBet         = NewGameError("cannot bet, must raise")
	ErrCannotRaise       = NewGameError("cannot raise, must bet")
	ErrBetTooSmall       = NewGameError("bet too small")
	ErrRaiseTooSmall     = NewGameError("raise too small")
	ErrInsufficientChips = NewGameError("insufficient chips")
	ErrNoChips           = NewGameError("no chips remaining")
	ErrNotYourTurn       = NewGameError("not your turn")
	ErrGameNotStarted    = NewGameError("game not started")
	ErrGameInProgress    = NewGameError("game already in progress")
	ErrTooFewPlayers     = NewGameError("too few players")
	ErrTooManyPlayers    = NewGameError("too many players")
	ErrPlayerNotFound    = NewGameError("player not found")
	ErrInvalidAction     = NewGameError("invalid action")
)

// GameError represents a game-specific error
type GameError struct {
	Message string
}

func NewGameError(message string) *GameError {
	return &GameError{Message: message}
}

func (e *GameError) Error() string {
	return e.Message
}

// Utility functions

// CalculatePotOdds calculates pot odds as a percentage
func CalculatePotOdds(callAmount, potSize int) float64 {
	if callAmount <= 0 {
		return 0
	}
	totalPot := potSize + callAmount
	return float64(callAmount) / float64(totalPot) * 100
}

// IsValidBetSize checks if a bet size is valid
func IsValidBetSize(betAmount, bigBlind, playerChips int) bool {
	if betAmount < bigBlind {
		return false
	}
	if betAmount > playerChips {
		return false
	}
	return true
}

// IsValidRaiseSize checks if a raise size is valid
func IsValidRaiseSize(raiseAmount, currentBet, minRaise, playerChips int) bool {
	if raiseAmount < currentBet+minRaise {
		return false
	}
	if raiseAmount > playerChips {
		return false
	}
	return true
}

// CalculateMainPot calculates the main pot when there are all-in players
func CalculateMainPot(players []*PokerPlayer) int {
	// Find the smallest all-in amount
	minAllIn := MaxBuyIn
	allInPlayers := 0

	for _, p := range players {
		if p.IsAllIn && p.TotalBetInHand < minAllIn {
			minAllIn = p.TotalBetInHand
			allInPlayers++
		}
	}

	if allInPlayers == 0 {
		// No all-in players, calculate total pot
		total := 0
		for _, p := range players {
			total += p.TotalBetInHand
		}
		return total
	}

	// Calculate main pot (everyone can contest this)
	mainPot := 0
	for _, p := range players {
		if p.TotalBetInHand >= minAllIn {
			mainPot += minAllIn
		} else {
			mainPot += p.TotalBetInHand
		}
	}

	return mainPot
}
