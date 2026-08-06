package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"germinaStack/model"
)

var ErrReactionNotFound = errors.New("reaction not found")

type ReactionResult struct {
	Reaction *model.ReactionType `json:"reacao"`
	Likes    int64               `json:"likes"`
	Dislikes int64               `json:"dislikes"`
}

type PostgresReactionRepository struct {
	db *sql.DB
}

func NewPostgresReactionRepository(db *sql.DB) *PostgresReactionRepository {
	return &PostgresReactionRepository{db: db}
}

func (r *PostgresReactionRepository) ToggleReaction(ctx context.Context, userID int64, messageType string, messageID int64, reaction model.ReactionType) (ReactionResult, error) {
	if !validReactionTarget(messageType) || !validReactionType(reaction) || userID <= 0 || messageID <= 0 {
		return ReactionResult{}, ErrReactionNotFound
	}
	if err := r.ensureTargetExists(ctx, messageType, messageID); err != nil {
		return ReactionResult{}, err
	}
	if _, err := r.db.ExecContext(ctx, "SELECT reaction($1, $2, $3, $4)", userID, messageID, messageType, reaction); err != nil {
		return ReactionResult{}, fmt.Errorf("toggle reaction: %w", err)
	}
	return r.readReactionResult(ctx, userID, messageType, messageID)
}

func (r *PostgresReactionRepository) ensureTargetExists(ctx context.Context, messageType string, messageID int64) error {
	table := map[string]string{"post": "posts", "comment": "comments", "comment_on_comment": "comments_on_comments"}[messageType]
	var id int64
	if err := r.db.QueryRowContext(ctx, "SELECT id FROM "+table+" WHERE id = $1", messageID).Scan(&id); errors.Is(err, sql.ErrNoRows) {
		return ErrReactionNotFound
	} else if err != nil {
		return fmt.Errorf("find reaction target: %w", err)
	}
	return nil
}

func (r *PostgresReactionRepository) GetReaction(ctx context.Context, userID int64, messageType string, messageID int64) (*model.ReactionType, error) {
	if !validReactionTarget(messageType) || userID <= 0 || messageID <= 0 {
		return nil, ErrReactionNotFound
	}
	query := `SELECT reaction_type FROM reactions WHERE id_user = $1 AND id_` + reactionTargetColumn(messageType) + ` = $2`
	var reaction model.ReactionType
	if err := r.db.QueryRowContext(ctx, query, userID, messageID).Scan(&reaction); errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("get reaction: %w", err)
	}
	return &reaction, nil
}

func (r *PostgresReactionRepository) readReactionResult(ctx context.Context, userID int64, messageType string, messageID int64) (ReactionResult, error) {
	reaction, err := r.GetReaction(ctx, userID, messageType, messageID)
	if err != nil {
		return ReactionResult{}, err
	}
	table := map[string]string{"post": "posts", "comment": "comments", "comment_on_comment": "comments_on_comments"}[messageType]
	var likes, dislikes int64
	query := `SELECT likes, dislikes FROM ` + table + ` WHERE id = $1`
	if err := r.db.QueryRowContext(ctx, query, messageID).Scan(&likes, &dislikes); err != nil {
		return ReactionResult{}, fmt.Errorf("count reactions: %w", err)
	}
	return ReactionResult{Reaction: reaction, Likes: likes, Dislikes: dislikes}, nil
}

func validReactionTarget(value string) bool {
	return value == "post" || value == "comment" || value == "comment_on_comment"
}

func validReactionType(value model.ReactionType) bool {
	return value == model.ReactionTypeLike || value == model.ReactionTypeDislike
}

func reactionTargetColumn(value string) string {
	switch value {
	case "post":
		return "post"
	case "comment":
		return "comment"
	default:
		return "comment_on_comment"
	}
}
