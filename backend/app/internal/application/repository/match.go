package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"game/api/internal/domain/entity"
	"game/api/internal/errs"
	"game/api/internal/infra/database"
	"game/api/internal/infra/logger"
)

const (
	matchKeyPrefix = "match:"
)

type Match struct {
	db    *database.Postgres
	cache *redis.Client
}

func NewMatch(db *database.Postgres, cache *redis.Client) *Match {
	return &Match{
		db:    db,
		cache: cache,
	}
}

func (r *Match) Create(ctx context.Context, match entity.Match) error {

	// Depois, inserir a partida
	matchData := database.MatchData{
		GUID:        match.ID().String(),
		GameMode:    string(match.GameMode()),
		MinPlayers:  match.MinPlayers(),
		MaxPlayers:  match.MaxPlayers(),
		Status:      string(match.Status()),
		Bets:        match.Bets(),
		Choices:     match.Choices(),
		Result:      match.Result(),
		CurrentTurn: match.CurrentTurn(),
	}

	if err := r.db.InsertMatch(ctx, matchData); err != nil {
		return err
	}
	// Primeiro, inserir o jogador na tabela match_players
	playerData := database.MatchPlayerData{
		GUID:     uuid.New().String(),
		MatchID:  match.ID().String(),
		PlayerID: match.CreatorID().String(),
		Role:     string(entity.PlayerRoleHost),
	}

	if err := r.db.InsertMatchPlayer(ctx, playerData); err != nil {
		return err
	}

	key := matchKeyPrefix + match.ID().String()
	data, err := json.Marshal(matchData)
	if err != nil {
		logger.Errorf("Failed to marshal match: %v", err)
		return err
	}

	if err := r.cache.Set(ctx, key, data, 0).Err(); err != nil {
		logger.Errorf("Failed to save match in cache: %v", err)
		return err
	}

	return nil
}

func (r *Match) GetMatchPlayers(ctx context.Context, matchID uuid.UUID) (players []entity.MatchPlayer, err error) {
	key := fmt.Sprintf("match:%s:players", matchID)
	var dataPlayers []database.MatchPlayerData = make([]database.MatchPlayerData, 0)
	dataJson, err := r.cache.Get(ctx, key).Bytes()
	if err != nil && err == redis.Nil {
		logger.Errorf("Failed to get match players from cache: %v", err)
		dataPlayers, err = r.db.GetMatchPlayers(ctx, matchID.String())
		if err != nil {
			return nil, err
		}

		// Serializar para JSON antes de armazenar no Redis
		jsonData, err := json.Marshal(dataPlayers)
		if err != nil {
			logger.Errorf("Failed to marshal match players: %v", err)
			return nil, err
		}

		if err := r.cache.Set(ctx, key, jsonData, 0).Err(); err != nil {
			logger.Errorf("Failed to save match players in cache: %v", err)
		}

	} else {
		if err := json.Unmarshal(dataJson, &dataPlayers); err != nil {
			logger.Errorf("Failed to unmarshal match players: %v", err)
			dataPlayers, err = r.db.GetMatchPlayers(ctx, matchID.String())
			if err != nil {
				return nil, err
			}
		}
	}
	matchPlayers := make([]entity.MatchPlayer, len(dataPlayers))
	for i, p := range dataPlayers {
		playerID, err := uuid.Parse(p.PlayerID)
		if err != nil {
			return nil, err
		}
		matchPlayers[i] = entity.MatchPlayer{
			PlayerID: playerID,
			Role:     entity.PlayerRole(p.Role),
		}
	}

	return matchPlayers, nil
}

func (r *Match) GetByID(ctx context.Context, id uuid.UUID) (m entity.Match, err error) {
	key := matchKeyPrefix + id.String()
	var data database.MatchData
	dataJson, err := r.cache.Get(ctx, key).Bytes()
	if err != nil && err == redis.Nil {
		logger.Errorf("Failed to get match from cache: %v", err)
		data, err = r.db.GetMatchByID(ctx, id.String())
	} else {
		if err = json.Unmarshal(dataJson, &data); err != nil {
			logger.Errorf("Failed to unmarshal match: %v", err)
			data, err = r.db.GetMatchByID(ctx, id.String())
		}
	}

	if err != nil {
		if err == sql.ErrNoRows {
			err = errs.ErrMatchNotFound
			return m, err
		}
		err = fmt.Errorf("failed to get match: %w", err)
		return m, err
	}

	// Busca os jogadores antes de montar a entidade
	players, err := r.GetMatchPlayers(ctx, id)
	if err != nil {
		return m, err
	}

	m, err = entity.LoadMatch(
		data.GUID,
		data.MinPlayers,
		data.MaxPlayers,
		entity.GameMode(data.GameMode),
		players,
		entity.MatchStatus(data.Status),
		data.CreatedAt,
		data.Choices,
		data.Bets,
		data.Result,
		data.CurrentTurn,
	)
	if err != nil {
		return m, err
	}

	// Atualiza o cache apenas se os dados vieram do banco
	if err == redis.Nil || err != nil {
		cacheData, err := json.Marshal(data)
		if err != nil {
			return m, err
		}

		if err := r.cache.Set(ctx, key, cacheData, 0).Err(); err != nil {
			return m, err
		}
	}

	return m, nil
}

func (r *Match) Update(ctx context.Context, match *entity.Match) error {
	matchData := database.MatchData{
		GUID:        match.ID().String(),
		GameMode:    string(match.GameMode()),
		MinPlayers:  match.MinPlayers(),
		MaxPlayers:  match.MaxPlayers(),
		Status:      string(match.Status()),
		Bets:        match.Bets(),
		Choices:     match.Choices(),
		Result:      match.Result(),
		CurrentTurn: match.CurrentTurn(),
	}

	if err := r.db.UpdateMatch(ctx, matchData); err != nil {
		return err
	}

	key := matchKeyPrefix + match.ID().String()
	data, err := json.Marshal(matchData)
	if err != nil {
		logger.Errorf("Failed to marshal match: %v", err)
		return err
	}

	if err := r.cache.Set(ctx, key, data, 0).Err(); err != nil {
		logger.Errorf("Failed to save match in cache: %v", err)
		return err
	}

	return nil
}

func (r *Match) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.DeleteMatch(ctx, id.String()); err != nil {
		return err
	}

	key := matchKeyPrefix + id.String()
	err := r.cache.Del(ctx, key).Err()
	if err != nil {
		logger.Errorf("Failed to delete match from cache: %v", err)
		return err
	}

	return nil
}

func (r *Match) Leave(ctx context.Context, matchID uuid.UUID, clientID uuid.UUID) error {
	if err := r.db.DeleteMatchPlayer(ctx, matchID.String(), clientID.String()); err != nil {
		return err
	}

	// Atualiza o cache
	match, err := r.GetByID(ctx, matchID)
	if err != nil {
		return err
	}

	if err := match.RemovePlayer(clientID); err != nil {
		return err
	}

	return r.Update(ctx, &match)
}

func (r *Match) Join(ctx context.Context, matchID uuid.UUID, clientID uuid.UUID) error {
	match, err := r.GetByID(ctx, matchID)
	if err != nil {
		return err
	}

	if err := match.AddPlayer(clientID); err != nil {
		return err
	}

	playerData := database.MatchPlayerData{
		GUID:     uuid.New().String(),
		MatchID:  matchID.String(),
		PlayerID: clientID.String(),
		Role:     string(entity.PlayerRoleGuest),
	}

	if err := r.db.InsertMatchPlayer(ctx, playerData); err != nil {
		return err
	}

	return r.Update(ctx, &match)
}
