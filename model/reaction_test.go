package model

import (
	"strings"
	"testing"
)

func int64Pointer(value int64) *int64 {
	return &value
}

func TestReactionValidate(t *testing.T) {
	tests := []struct {
		name         string
		reaction     Reaction
		errorContext string
	}{
		{
			name: "post com like",
			reaction: Reaction{
				PostID:       int64Pointer(1),
				ReactionType: ReactionTypeLike,
			},
		},
		{
			name: "comentário com dislike",
			reaction: Reaction{
				CommentID:    int64Pointer(1),
				ReactionType: ReactionTypeDislike,
			},
		},
		{
			name: "resposta com like",
			reaction: Reaction{
				CommentOnCommentID: int64Pointer(1),
				ReactionType:       ReactionTypeLike,
			},
		},
		{
			name: "alvo com ID zero",
			reaction: Reaction{
				PostID:       int64Pointer(0),
				ReactionType: ReactionTypeLike,
			},
		},
		{
			name: "sem alvo",
			reaction: Reaction{
				ReactionType: ReactionTypeLike,
			},
			errorContext: "exatamente um alvo",
		},
		{
			name: "dois alvos",
			reaction: Reaction{
				PostID:       int64Pointer(1),
				CommentID:    int64Pointer(2),
				ReactionType: ReactionTypeLike,
			},
			errorContext: "exatamente um alvo",
		},
		{
			name: "três alvos",
			reaction: Reaction{
				PostID:             int64Pointer(1),
				CommentID:          int64Pointer(2),
				CommentOnCommentID: int64Pointer(3),
				ReactionType:       ReactionTypeLike,
			},
			errorContext: "exatamente um alvo",
		},
		{
			name: "tipo inválido",
			reaction: Reaction{
				PostID:       int64Pointer(1),
				ReactionType: ReactionType("amei"),
			},
			errorContext: "tipo de reação",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.reaction.Validate()
			if test.errorContext == "" {
				if err != nil {
					t.Fatalf("Validate() error = %v, esperado nil", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("Validate() deveria rejeitar %s", test.errorContext)
			}
			if !strings.Contains(err.Error(), test.errorContext) {
				t.Fatalf("Validate() retornou erro pouco claro: %q", err)
			}
			if test.errorContext == "tipo de reação" &&
				!strings.Contains(err.Error(), "amei") {
				t.Fatalf("Validate() não informou o valor inválido: %q", err)
			}
		})
	}
}

func TestReactionTypeRecognizesSQLValues(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{value: "like", want: true},
		{value: "dislike", want: true},
		{value: ""},
		{value: "amei"},
	}

	for _, test := range tests {
		t.Run(test.value, func(t *testing.T) {
			if got := ReactionType(test.value).IsValid(); got != test.want {
				t.Fatalf("IsValid() = %v, esperado %v", got, test.want)
			}
		})
	}
}
