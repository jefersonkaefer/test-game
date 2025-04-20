package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"game/api/internal/errs"
	"game/api/internal/infra/logger"

	"github.com/sirupsen/logrus"
)

const (
	DB_TABLE_MATCHES       = "matches"
	DB_TABLE_MATCH_PLAYERS = "match_players"
)

type MatchData struct {
	GUID        string             `db:"guid"`
	GameMode    string             `db:"game_mode"`
	MinPlayers  int                `db:"min_players"`
	MaxPlayers  int                `db:"max_players"`
	Status      string             `db:"status"`
	Bets        map[string]float64 `db:"bets"`
	Choices     map[string]string  `db:"choices"`
	Result      string             `db:"result"`
	CurrentTurn string             `db:"current_turn"`
	CreatedAt   time.Time          `db:"created_at"`
	UpdatedAt   time.Time          `db:"updated_at"`
	DeletedAt   *time.Time         `db:"deleted_at"`
}

type MatchPlayerData struct {
	GUID      string     `db:"guid"`
	MatchID   string     `db:"match_id"`
	PlayerID  string     `db:"player_id"`
	Role      string     `db:"role"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

func (pg *Postgres) InsertMatch(ctx context.Context, data MatchData) error {
	logger.WithFields(logrus.Fields{
		"match_id": data.GUID,
	}).Debug("Inserting new match")

	betsJson, err := json.Marshal(data.Bets)
	if err != nil {
		return fmt.Errorf("failed to marshal bets: %w", err)
	}

	choicesJson, err := json.Marshal(data.Choices)
	if err != nil {
		return fmt.Errorf("failed to marshal choices: %w", err)
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (guid, game_mode, min_players, max_players, status, bets, choices, result, current_turn) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
		DB_TABLE_MATCHES,
	)
	_, err = pg.conn.Exec(
		query,
		data.GUID,
		data.GameMode,
		data.MinPlayers,
		data.MaxPlayers,
		data.Status,
		betsJson,
		choicesJson,
		data.Result,
		data.CurrentTurn,
	)
	if err != nil {
		logger.Errorf("Failed to insert match: %v", err)
		return err
	}

	logger.WithFields(logrus.Fields{
		"match_id": data.GUID,
	}).Info("Match inserted successfully")
	return nil
}

func (pg *Postgres) GetMatchByID(ctx context.Context, id string) (MatchData, error) {
	logger.WithFields(logrus.Fields{
		"match_id": id,
	}).Debug("Getting match by ID")

	var data MatchData
	var betsJson, choicesJson []byte

	err := pg.conn.QueryRow(
		fmt.Sprintf("SELECT guid, game_mode, min_players, max_players, status, bets, choices, result, current_turn, created_at, updated_at, deleted_at FROM %s WHERE guid = $1", DB_TABLE_MATCHES),
		id,
	).Scan(
		&data.GUID,
		&data.GameMode,
		&data.MinPlayers,
		&data.MaxPlayers,
		&data.Status,
		&betsJson,
		&choicesJson,
		&data.Result,
		&data.CurrentTurn,
		&data.CreatedAt,
		&data.UpdatedAt,
		&data.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.WithFields(logrus.Fields{
				"match_id": id,
			}).Warn("Match not found")
			return MatchData{}, errs.ErrMatchNotFound
		}
		logger.Errorf("Failed to get match: %v", err)
		return MatchData{}, err
	}

	if err := json.Unmarshal(betsJson, &data.Bets); err != nil {
		return MatchData{}, fmt.Errorf("failed to unmarshal bets: %w", err)
	}

	if err := json.Unmarshal(choicesJson, &data.Choices); err != nil {
		return MatchData{}, fmt.Errorf("failed to unmarshal choices: %w", err)
	}

	return data, nil
}

func (pg *Postgres) UpdateMatch(ctx context.Context, data MatchData) error {
	logger.WithFields(logrus.Fields{
		"match_id": data.GUID,
	}).Debug("Updating match")

	betsJson, err := json.Marshal(data.Bets)
	if err != nil {
		return fmt.Errorf("failed to marshal bets: %w", err)
	}

	choicesJson, err := json.Marshal(data.Choices)
	if err != nil {
		return fmt.Errorf("failed to marshal choices: %w", err)
	}

	query := fmt.Sprintf(
		"UPDATE %s SET status = $1, bets = $2, choices = $3, result = $4, current_turn = $5, updated_at = NOW() WHERE guid = $6",
		DB_TABLE_MATCHES,
	)
	_, err = pg.conn.Exec(query, data.Status, betsJson, choicesJson, data.Result, data.CurrentTurn, data.GUID)
	if err != nil {
		logger.Errorf("Failed to update match: %v", err)
		return err
	}

	logger.WithFields(logrus.Fields{
		"match_id": data.GUID,
	}).Info("Match updated successfully")
	return nil
}

func (pg *Postgres) DeleteMatch(ctx context.Context, id string) error {
	logger.WithFields(logrus.Fields{
		"match_id": id,
	}).Debug("Deleting match")

	query := fmt.Sprintf("UPDATE %s SET deleted_at = NOW() WHERE guid = $1", DB_TABLE_MATCHES)
	_, err := pg.conn.Exec(query, id)
	if err != nil {
		logger.Errorf("Failed to delete match: %v", err)
		return err
	}

	logger.WithFields(logrus.Fields{
		"match_id": id,
	}).Info("Match deleted successfully")
	return nil
}

func (pg *Postgres) InsertMatchPlayer(ctx context.Context, data MatchPlayerData) error {
	logger.WithFields(logrus.Fields{
		"match_id":  data.MatchID,
		"player_id": data.PlayerID,
	}).Debug("Inserting match player")

	query := fmt.Sprintf(
		"INSERT INTO %s (guid, match_id, player_id, role) VALUES ($1, $2, $3, $4)",
		DB_TABLE_MATCH_PLAYERS,
	)
	_, err := pg.conn.Exec(
		query,
		data.GUID,
		data.MatchID,
		data.PlayerID,
		data.Role,
	)
	if err != nil {
		logger.Errorf("Failed to insert match player: %v", err)
		return err
	}

	return nil
}

func (pg *Postgres) GetMatchPlayers(ctx context.Context, matchID string) ([]MatchPlayerData, error) {
	logger.WithFields(logrus.Fields{
		"match_id": matchID,
	}).Debug("Getting match players")

	query := fmt.Sprintf("SELECT * FROM %s WHERE match_id = $1 AND deleted_at IS NULL", DB_TABLE_MATCH_PLAYERS)
	rows, err := pg.conn.Query(query, matchID)
	if err != nil {
		logger.Errorf("Failed to get match players: %v", err)
		return nil, err
	}
	defer rows.Close()

	var players []MatchPlayerData
	for rows.Next() {
		var player MatchPlayerData
		err := rows.Scan(
			&player.GUID,
			&player.MatchID,
			&player.PlayerID,
			&player.Role,
			&player.CreatedAt,
			&player.UpdatedAt,
			&player.DeletedAt,
		)
		if err != nil {
			logger.Errorf("Failed to scan match player: %v", err)
			return nil, err
		}
		players = append(players, player)
	}

	return players, nil
}

func (pg *Postgres) DeleteMatchPlayer(ctx context.Context, matchID, playerID string) error {
	logger.WithFields(logrus.Fields{
		"match_id":  matchID,
		"player_id": playerID,
	}).Debug("Deleting match player")

	query := fmt.Sprintf("UPDATE %s SET deleted_at = NOW() WHERE match_id = $1 AND player_id = $2", DB_TABLE_MATCH_PLAYERS)
	_, err := pg.conn.Exec(query, matchID, playerID)
	if err != nil {
		logger.Errorf("Failed to delete match player: %v", err)
		return err
	}

	return nil
}
