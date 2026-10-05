package services

import (
	"context"
	"errors"
	"fmt"
	"lotoMironBot/internal/database"
	"lotoMironBot/internal/repository"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type LotoService struct {
	repo *repository.LotoRepository
}

var (
	ErrActiveGameExists    = errors.New("administrator already has an active game")
	ErrChatGameExists      = errors.New("group already has an active game")
	ErrGameNotFound        = errors.New("active game was not found")
	ErrRecruitmentClosed   = errors.New("recruitment is closed")
	ErrParticipantLimit    = errors.New("participant limit has been reached")
	ErrGameNotReady        = errors.New("game does not have the required participants and tickets")
	ErrGameNotRunning      = errors.New("game is not running")
	ErrDeckExhausted       = errors.New("all barrels have already been drawn")
	ErrInvalidGiveCount    = errors.New("ticket quantity must be from 1 to 5")
	ErrExistingTicketCount = errors.New("player already has more tickets than requested")
	ErrClaimWindowOpen     = errors.New("the current number is still available for 15 seconds")
	ErrClaimWindowExpired  = errors.New("the 15-second answer window has expired")
	ErrNotGameParticipant  = errors.New("player is not an active participant")
	ErrAlreadyAttempted    = errors.New("player already answered for this number")
)

type AssignedTicket struct {
	PlayerID int64
	Ticket   *database.Ticket
}

type DrawResult struct {
	GameID          uint
	Number          int32
	DrawIndex       int
	WinnerPlayerIDs []int64
	Finished        bool
}

type ClaimResult struct {
	Success       bool
	Failures      int
	Disqualified  bool
	Winner        bool
	GameFinished  bool
	ClosedTickets int
}

type PlayerGameStat struct {
	PlayerID     int64
	Tickets      int
	Closed       int
	Failures     int
	Disqualified bool
	Winner       bool
}

func allTicketsClosed(total, closed int) bool {
	return total > 0 && closed == total
}

func reachedMissLimit(failures int) bool {
	return failures >= 9
}

func isTicketClosed(ticket *database.GameTicket) bool {
	return ticket.Completed || len(ticket.Numbers) == 0 || len(ticket.MarkedNumbers) >= 6
}

const DefaultRules = `🎟 ПРАВИЛА ИГРЫ В ЛОТО

• Можно приобрести до 5 билетов.
• В каждом билете — 6 случайных чисел от 1 до 100.
• Выигрышный билет — тот, в котором зачёркнуты все 6 чисел.
• Игра проходит в чате, где были приобретены билеты.

⏱ 15 секунд на реакцию
После выпадения вашего числа у вас есть 15 секунд, чтобы нажать кнопку «У меня есть».

⚠️ Если вы пропустили своё число и не нажали кнопку вовремя — билет считается проигравшим.

❌ Ошибочные нажатия
Если вы нажали «У меня есть» на число, которого нет ни в одном вашем билете, это считается ошибкой.
У каждого игрока есть 9 ошибок. После 9 ошибочных нажатий происходит автоматическая дисквалификация.

Следите за игрой внимательно! Заранее выпишите числа со своих билетов.

🏆 Игра автоматически останавливается, когда у одного из игроков закрыты все его билеты.`

func NewLotoService(repo *repository.LotoRepository) *LotoService {
	return &LotoService{
		repo: repo,
	}
}

func (s *LotoService) AddTicket(ctx context.Context, fileID, collection string, numbers []int32) (*database.Ticket, error) {
	ticket := &database.Ticket{
		FileID:     fileID,
		Collection: database.Collections(collection),
		Numbers:    pq.Int32Array(numbers),
	}

	addedTicket, err := s.repo.AddTicket(ctx, ticket)
	if err != nil {
		return nil, err
	}

	return addedTicket, nil
}

func (s *LotoService) GetAllTickets(ctx context.Context) ([]*database.Ticket, error) {
	return s.repo.GetAllTickets(ctx)
}

func (s *LotoService) GetRules(ctx context.Context) (string, error) {
	rules, err := s.repo.GetSettings(ctx, "game_rules")
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return DefaultRules, nil
	}
	return rules, err
}

func (s *LotoService) SetRules(ctx context.Context, rules string) error {
	if strings.TrimSpace(rules) == "" {
		return fmt.Errorf("rules cannot be empty")
	}
	return s.repo.SaveSettings(ctx, "game_rules", rules)
}

func (s *LotoService) GetDefaultCollection(ctx context.Context) (database.Collections, error) {
	collection, err := s.repo.GetSettings(ctx, "default_game_collection")
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return database.Standard, nil
	}
	return database.Collections(collection), err
}

func (s *LotoService) SetDefaultCollection(ctx context.Context, collection database.Collections) error {
	switch collection {
	case database.Standard, database.NewYear, database.Halloween:
		return s.repo.SaveSettings(ctx, "default_game_collection", string(collection))
	default:
		return fmt.Errorf("invalid game collection: %s", collection)
	}
}

func (s *LotoService) GetPlayerGameStats(ctx context.Context, chatID int64) ([]PlayerGameStat, error) {
	_, participants, tickets, err := s.repo.ListLatestGameStats(ctx, chatID)
	if err != nil {
		return nil, err
	}
	stats := make([]PlayerGameStat, 0, len(participants))
	byID := make(map[int64]int, len(participants))
	for _, p := range participants {
		byID[p.PlayerID] = len(stats)
		stats = append(stats, PlayerGameStat{PlayerID: p.PlayerID, Failures: p.Failures, Disqualified: p.Disqualified})
	}
	for _, ticket := range tickets {
		i, ok := byID[ticket.PlayerID]
		if !ok {
			continue
		}
		stats[i].Tickets++
		if isTicketClosed(ticket) {
			stats[i].Closed++
		}
	}
	for i := range stats {
		stats[i].Winner = stats[i].Tickets > 0 && stats[i].Tickets == stats[i].Closed
	}
	return stats, nil
}

func (s *LotoService) GetGameWinners(ctx context.Context, chatID int64) ([]PlayerGameStat, error) {
	_, participants, tickets, err := s.repo.ListLatestGameStats(ctx, chatID)
	if err != nil {
		return nil, err
	}
	byID := make(map[int64]PlayerGameStat, len(participants))
	for _, p := range participants {
		byID[p.PlayerID] = PlayerGameStat{PlayerID: p.PlayerID, Failures: p.Failures, Disqualified: p.Disqualified}
	}
	for _, ticket := range tickets {
		stat := byID[ticket.PlayerID]
		stat.Tickets++
		if isTicketClosed(ticket) {
			stat.Closed++
		}
		byID[ticket.PlayerID] = stat
	}
	var winners []PlayerGameStat
	for _, stat := range byID {
		if stat.Tickets > 0 && stat.Tickets == stat.Closed {
			stat.Winner = true
			winners = append(winners, stat)
		}
	}
	return winners, nil
}

func (s *LotoService) GetPlayerTicketsPrivate(ctx context.Context, playerID int64) ([]*database.Ticket, error) {
	gameTickets, err := s.repo.ListLatestPlayerGameTickets(ctx, playerID)
	if err != nil {
		return nil, err
	}
	tickets := make([]*database.Ticket, 0, len(gameTickets))
	for _, gameTicket := range gameTickets {
		ticket, err := s.repo.GetTicketByID(ctx, gameTicket.TicketID)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, ticket)
	}
	return tickets, nil
}

func (s *LotoService) CreateGame(ctx context.Context, adminID, chatID int64, participantLimit int, collection database.Collections) (*database.Game, error) {
	if collection != database.Standard && collection != database.NewYear && collection != database.Halloween {
		return nil, fmt.Errorf("unknown ticket collection")
	}

	var game *database.Game
	err := s.repo.WithinTransaction(ctx, func(repo *repository.LotoRepository) error {
		if err := repo.LockAdmin(ctx, adminID); err != nil {
			return err
		}
		if err := repo.LockChat(ctx, chatID); err != nil {
			return err
		}
		if _, err := repo.FindActiveGameByAdmin(ctx, adminID); err == nil {
			return ErrActiveGameExists
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if _, err := repo.FindActiveGameByChat(ctx, chatID); err == nil {
			return ErrChatGameExists
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		deck := make(pq.Int32Array, 100)
		for i := range deck {
			deck[i] = int32(i + 1)
		}
		rand.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })

		created, err := repo.CreateGame(ctx, &database.Game{
			AdminID:          adminID,
			ChatID:           chatID,
			Collection:       collection,
			Status:           database.GamePending,
			ParticipantLimit: participantLimit,
			TicketsPerPlayer: 1,
			Deck:             deck,
			DrawIndex:        0,
		})
		if err != nil {
			return err
		}
		game = created
		return nil
	})
	return game, err
}

func (s *LotoService) GiveTickets(ctx context.Context, chatID, playerID int64, desiredCount int) ([]AssignedTicket, error) {
	if desiredCount < 1 || desiredCount > 5 {
		return nil, ErrInvalidGiveCount
	}
	var assignments []AssignedTicket
	err := s.repo.WithinTransaction(ctx, func(repo *repository.LotoRepository) error {
		game, err := repo.LockActiveGameByChat(ctx, chatID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGameNotFound
		}
		if err != nil {
			return err
		}
		if game.Status != database.GamePending {
			return ErrRecruitmentClosed
		}

		registered, err := repo.IsGameParticipant(ctx, game.ID, playerID)
		if err != nil {
			return err
		}
		if !registered {
			if err := repo.CreateGameParticipant(ctx, &database.GameParticipant{GameID: game.ID, PlayerID: playerID}); err != nil {
				return err
			}
		}

		assigned, err := repo.ListPlayerGameTickets(ctx, game.ID, playerID)
		if err != nil {
			return err
		}
		if len(assigned) > desiredCount {
			return ErrExistingTicketCount
		}
		missing := desiredCount - len(assigned)
		available, err := repo.ListAvailableTickets(ctx, game.ID, game.Collection, missing)
		if err != nil {
			return err
		}
		if len(available) != missing {
			return fmt.Errorf("not enough available tickets in collection")
		}
		rand.Shuffle(len(available), func(i, j int) { available[i], available[j] = available[j], available[i] })
		for _, ticket := range available {
			if _, err := repo.CreateGameTicket(ctx, &database.GameTicket{
				GameID: game.ID, PlayerID: playerID, TicketID: ticket.ID,
				Numbers: append(pq.Int32Array(nil), ticket.Numbers...),
			}); err != nil {
				return err
			}
		}

		assigned, err = repo.ListPlayerGameTickets(ctx, game.ID, playerID)
		if err != nil {
			return err
		}
		for _, gameTicket := range assigned {
			ticket, err := repo.GetTicketByID(ctx, gameTicket.TicketID)
			if err != nil {
				return err
			}
			assignments = append(assignments, AssignedTicket{PlayerID: playerID, Ticket: ticket})
		}
		return nil
	})
	return assignments, err
}

func (s *LotoService) GetPlayerTickets(ctx context.Context, chatID, playerID int64) ([]*database.Ticket, error) {
	game, err := s.repo.FindActiveGameByChat(ctx, chatID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrGameNotFound
	}
	if err != nil {
		return nil, err
	}
	if game.Status != database.GameStarted {
		return nil, ErrGameNotRunning
	}
	registered, err := s.repo.IsGameParticipant(ctx, game.ID, playerID)
	if err != nil {
		return nil, err
	}
	if !registered {
		return nil, ErrNotGameParticipant
	}
	gameTickets, err := s.repo.ListPlayerGameTickets(ctx, game.ID, playerID)
	if err != nil {
		return nil, err
	}
	tickets := make([]*database.Ticket, 0, len(gameTickets))
	for _, gameTicket := range gameTickets {
		ticket, err := s.repo.GetTicketByID(ctx, gameTicket.TicketID)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, ticket)
	}
	return tickets, nil
}

func (s *LotoService) StartGame(ctx context.Context, chatID int64) error {
	return s.repo.WithinTransaction(ctx, func(repo *repository.LotoRepository) error {
		game, err := repo.LockActiveGameByChat(ctx, chatID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGameNotFound
		}
		if err != nil {
			return err
		}
		if game.Status != database.GamePending {
			return ErrGameNotReady
		}
		participantCount, err := repo.CountGameParticipants(ctx, game.ID)
		if err != nil {
			return err
		}
		if participantCount == 0 {
			return ErrGameNotReady
		}
		eligibleCount, err := repo.CountEligibleParticipants(ctx, game.ID, game.AdminID)
		if err != nil {
			return err
		}
		if eligibleCount == 0 {
			return ErrGameNotReady
		}
		participants, err := repo.ListGameParticipants(ctx, game.ID)
		if err != nil {
			return err
		}
		for _, participant := range participants {
			count, err := repo.CountPlayerGameTickets(ctx, game.ID, participant.PlayerID)
			if err != nil || count == 0 || count > 5 {
				return ErrGameNotReady
			}
		}
		now := time.Now()
		game.Status = database.GameStarted
		game.StartedAt = &now
		return repo.SaveGame(ctx, game)
	})
}

func (s *LotoService) Draw(ctx context.Context, chatID int64) (*DrawResult, error) {
	var result DrawResult
	err := s.repo.WithinTransaction(ctx, func(repo *repository.LotoRepository) error {
		game, err := repo.LockActiveGameByChat(ctx, chatID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGameNotFound
		}
		if err != nil {
			return err
		}
		if game.Status != database.GameStarted {
			return ErrGameNotRunning
		}
		if game.DrawIndex >= len(game.Deck) {
			now := time.Now()
			game.Status = database.GameFinished
			game.FinishedAt = &now
			result.GameID = game.ID
			result.Finished = true
			return repo.SaveGame(ctx, game)
		}
		if game.LastDrawAt != nil && time.Since(*game.LastDrawAt) < 15*time.Second {
			return ErrClaimWindowOpen
		}
		result.Number = game.Deck[game.DrawIndex]
		result.GameID = game.ID
		game.DrawIndex++
		now := time.Now()
		game.LastDrawAt = &now
		result.DrawIndex = game.DrawIndex
		return repo.SaveGame(ctx, game)
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *LotoService) ClaimDrawNumber(ctx context.Context, chatID int64, gameID uint, playerID int64, drawnNumber int32, drawIndex int) (*ClaimResult, error) {
	result := &ClaimResult{}
	err := s.repo.WithinTransaction(ctx, func(repo *repository.LotoRepository) error {
		game, err := repo.LockActiveGameByChat(ctx, chatID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGameNotFound
		}
		if err != nil {
			return err
		}
		if game.ID != gameID || game.Status != database.GameStarted {
			return ErrGameNotRunning
		}
		participant, err := repo.GetGameParticipant(ctx, gameID, playerID)
		if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && participant.Disqualified) {
			return ErrNotGameParticipant
		}
		if err != nil {
			return err
		}
		if drawIndex != game.DrawIndex || game.DrawIndex == 0 || game.LastDrawAt == nil || game.Deck[drawIndex-1] != drawnNumber {
			return ErrClaimWindowExpired
		}
		if time.Since(*game.LastDrawAt) > 15*time.Second {
			return ErrClaimWindowExpired
		}
		if participant.LastAttemptDrawIndex == drawIndex {
			return ErrAlreadyAttempted
		}
		participant.LastAttemptDrawIndex = drawIndex

		tickets, err := repo.GetGameTickets(ctx, gameID)
		if err != nil {
			return err
		}
		matched := false
		for _, ticket := range tickets {
			if ticket.PlayerID != playerID || ticket.Completed || !contains(ticket.Numbers, drawnNumber) {
				continue
			}
			matched = true
			ticket.Numbers = without(ticket.Numbers, drawnNumber)
			ticket.MarkedNumbers = append(ticket.MarkedNumbers, drawnNumber)
			if len(ticket.Numbers) == 0 {
				ticket.Completed = true
			}
			if err := repo.SaveGameTicket(ctx, ticket); err != nil {
				return err
			}
		}
		if matched {
			playerTickets, err := repo.ListPlayerGameTickets(ctx, gameID, playerID)
			if err != nil {
				return err
			}
			closed := 0
			for _, ticket := range playerTickets {
				if isTicketClosed(ticket) {
					closed++
				}
			}
			result.ClosedTickets = closed
			result.Winner = allTicketsClosed(len(playerTickets), closed)
		}
		if matched {
			result.Success = true
		} else {
			participant.Failures++
			result.Failures = participant.Failures
			if reachedMissLimit(participant.Failures) {
				participant.Disqualified = true
				result.Disqualified = true
			}
		}
		if err := repo.SaveGameParticipant(ctx, participant); err != nil {
			return err
		}

		if result.Winner {
			now := time.Now()
			game.Status = database.GameFinished
			game.FinishedAt = &now
			game.WinnerPlayerID = &playerID
			game.WinnerTicketCount = result.ClosedTickets
			result.GameFinished = true
		} else if result.Disqualified {
			remaining, err := repo.CountEligibleParticipants(ctx, gameID, game.AdminID)
			if err != nil {
				return err
			}
			if remaining == 0 {
				now := time.Now()
				game.Status = database.GameFinished
				game.FinishedAt = &now
				result.GameFinished = true
			}
		}
		return repo.SaveGame(ctx, game)
	})
	return result, err
}

func (s *LotoService) EndGame(ctx context.Context, chatID int64) error {
	return s.repo.WithinTransaction(ctx, func(repo *repository.LotoRepository) error {
		game, err := repo.LockActiveGameByChat(ctx, chatID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrGameNotFound
		}
		if err != nil {
			return err
		}
		now := time.Now()
		game.Status = database.GameCanceled
		game.CanceledAt = &now
		game.FinishedAt = &now
		return repo.SaveGame(ctx, game)
	})
}

func (s *LotoService) IsCurrentDraw(ctx context.Context, chatID int64, gameID uint, drawIndex int) bool {
	game, err := s.repo.FindActiveGameByChat(ctx, chatID)
	return err == nil && game.ID == gameID && game.Status == database.GameStarted && game.DrawIndex == drawIndex
}

func contains(numbers pq.Int32Array, number int32) bool {
	for _, value := range numbers {
		if value == number {
			return true
		}
	}
	return false
}

func without(numbers pq.Int32Array, number int32) pq.Int32Array {
	result := make(pq.Int32Array, 0, len(numbers)-1)
	for _, value := range numbers {
		if value != number {
			result = append(result, value)
		}
	}
	return result
}
