package model

import (
	"errors"
	"fmt"
	"time"
)

type ReactionType string

const (
	ReactionTypeLike    ReactionType = "like"
	ReactionTypeDislike ReactionType = "dislike"
)

func (reactionType ReactionType) IsValid() bool {
	switch reactionType {
	case ReactionTypeLike, ReactionTypeDislike:
		return true
	default:
		return false
	}
}

type Reaction struct {
	ID                 int64        `db:"id" json:"id"`
	UserID             int64        `db:"id_user" json:"id_user"`
	PostID             *int64       `db:"id_post" json:"id_post"`
	CommentID          *int64       `db:"id_comment" json:"id_comment"`
	CommentOnCommentID *int64       `db:"id_comment_on_comment" json:"id_comment_on_comment"`
	ReactionType       ReactionType `db:"reaction_type" json:"reaction_type"`
	CreatedAt          time.Time    `db:"created_at" json:"created_at"`
}

func (reaction Reaction) Validate() error {
	targetCount := 0
	for _, target := range []*int64{
		reaction.PostID,
		reaction.CommentID,
		reaction.CommentOnCommentID,
	} {
		if target != nil {
			targetCount++
		}
	}

	if targetCount != 1 {
		return errors.New("a reação deve apontar para exatamente um alvo")
	}
	if !reaction.ReactionType.IsValid() {
		return fmt.Errorf("tipo de reação inválido: %q", reaction.ReactionType)
	}

	return nil
}
