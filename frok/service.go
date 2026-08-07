package frok

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"germinaStack/model"
)

var mentionPattern = regexp.MustCompile(`(?i)(^|[^[:alnum:]_])@frok($|[^[:alnum:]_])`)

type Repository interface {
	BotUserID(context.Context) (int64, error)
	Username(context.Context, int64) (string, error)
	GetPost(context.Context, int64) (model.Post, error)
	GetComment(context.Context, int64) (model.Comment, error)
	ListReplies(context.Context, int64) ([]model.CommentOnComment, error)
	CreateComment(context.Context, int64, int64, string) error
	CreateReply(context.Context, int64, int64, string) error
}

type ReplyClient interface {
	Reply(context.Context, string) (string, error)
}

type Memory struct {
	Prompt string
	Reply  string
}

type MemoryStore interface {
	Recall(context.Context, int64) ([]Memory, error)
	Remember(context.Context, int64, string, string) error
}

type NoopMemoryStore struct{}

func (NoopMemoryStore) Recall(context.Context, int64) ([]Memory, error) { return nil, nil }
func (NoopMemoryStore) Remember(context.Context, int64, string, string) error {
	return nil
}

type Service struct {
	repository Repository
	client     ReplyClient
	memory     MemoryStore
	timeout    time.Duration
	logError   func(error)
}

func NewService(repository Repository, client ReplyClient, timeout time.Duration, logError func(error), memories ...MemoryStore) *Service {
	memory := MemoryStore(NoopMemoryStore{})
	if len(memories) > 0 && memories[0] != nil {
		memory = memories[0]
	}
	return &Service{repository: repository, client: client, memory: memory, timeout: timeout, logError: logError}
}

func IsMentioned(content string) bool {
	return mentionPattern.MatchString(content)
}

func (s *Service) DispatchPost(authorID int64, post model.Post) {
	if !IsMentioned(post.Title + "\n" + post.Content) {
		return
	}
	go s.respond(authorID, post.ID, 0, postContext(post), postMemory(post))
}

func (s *Service) DispatchComment(authorID int64, comment model.Comment) {
	if !IsMentioned(comment.Content) {
		return
	}
	go s.respondToThread(authorID, comment.ID, comment.Content)
}

func (s *Service) DispatchReply(authorID int64, reply model.CommentOnComment) {
	if !IsMentioned(reply.Content) && !s.repliedInThread(reply.CommentID) {
		return
	}
	go s.respondToThread(authorID, reply.CommentID, reply.Content)
}

func (s *Service) repliedInThread(commentID int64) bool {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	botID, err := s.repository.BotUserID(ctx)
	if err != nil {
		s.report(err)
		return false
	}
	comment, err := s.repository.GetComment(ctx, commentID)
	if err != nil {
		s.report(err)
		return false
	}
	if comment.UserID == botID {
		return true
	}
	replies, err := s.repository.ListReplies(ctx, commentID)
	if err != nil {
		s.report(err)
		return false
	}
	for _, item := range replies {
		if item.UserID == botID {
			return true
		}
	}
	return false
}

func (s *Service) respond(authorID, postID, commentID int64, input, memoryPrompt string) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	s.respondWithContext(ctx, authorID, postID, commentID, input, memoryPrompt)
}

func (s *Service) respondToThread(authorID, commentID int64, memoryPrompt string) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	comment, err := s.repository.GetComment(ctx, commentID)
	if err != nil {
		s.report(err)
		return
	}
	post, err := s.repository.GetPost(ctx, comment.PostID)
	if err != nil {
		s.report(err)
		return
	}
	replies, err := s.repository.ListReplies(ctx, commentID)
	if err != nil {
		s.report(err)
		return
	}
	s.respondWithContext(ctx, authorID, 0, commentID, threadContext(post, comment, replies), memoryPrompt)
}

func (s *Service) respondWithContext(ctx context.Context, authorID, postID, commentID int64, input, memoryPrompt string) {
	username, err := s.repository.Username(ctx, authorID)
	if err != nil {
		s.report(err)
		return
	}
	memories, err := s.memory.Recall(ctx, authorID)
	if err != nil {
		s.report(err)
	}
	reply, err := s.client.Reply(ctx, withMemory(input, memories))
	if err != nil {
		s.report(err)
		return
	}
	botID, err := s.repository.BotUserID(ctx)
	if err != nil {
		s.report(err)
		return
	}
	content := formatReply(username, reply)
	if postID > 0 {
		err = s.repository.CreateComment(ctx, botID, postID, content)
	} else {
		err = s.repository.CreateReply(ctx, botID, commentID, content)
	}
	if err != nil {
		s.report(err)
		return
	}
	if err := s.memory.Remember(ctx, authorID, memoryPrompt, reply); err != nil {
		s.report(err)
	}
}

func postMemory(post model.Post) string {
	return postContext(post)
}

func withMemory(input string, memories []Memory) string {
	if len(memories) == 0 {
		return input
	}
	context := "Memórias anteriores deste mesmo usuário:"
	for _, memory := range memories {
		context += "\n- Pergunta: " + shortenMemory(memory.Prompt) + "\n  Resposta do Frok: " + shortenMemory(memory.Reply)
	}
	return context + "\n\nContexto atual:\n" + input
}

func shortenMemory(value string) string {
	const maxRunes = 1500
	runes := []rune(strings.TrimSpace(value))
	if len(runes) <= maxRunes {
		return string(runes)
	}
	// ponytail: cada memória é limitada para manter o prompt do Groq previsível; usar orçamento por tokens se crescer.
	return string(runes[:maxRunes]) + "…"
}

func formatReply(username, reply string) string {
	reply = strings.ToValidUTF8(reply, "")
	reply = strings.NewReplacer("@", "＠", "\\", "", "`", "", "*", "").Replace(reply)
	if !utf8.ValidString(reply) {
		reply = "Não consegui montar uma resposta em texto simples."
	}
	reply = strings.TrimSpace(reply)
	return "@" + username + " " + reply
}

func postContext(post model.Post) string {
	context := fmt.Sprintf("Título: %s\n\nPost: %s", post.Title, post.Content)
	if post.ImageDescription != nil && strings.TrimSpace(*post.ImageDescription) != "" {
		context += "\n\nDescrição da imagem (alt): " + strings.TrimSpace(*post.ImageDescription)
	}
	return context
}

func threadContext(post model.Post, comment model.Comment, replies []model.CommentOnComment) string {
	context := postContext(post) + "\n\nComentário relacionado de " + authorLabel(comment.AuthorName, comment.AuthorUsername) + ": " + comment.Content
	if len(replies) == 0 {
		return context
	}
	context += "\n\nRespostas relacionadas:"
	for _, reply := range replies {
		context += "\n- " + authorLabel(reply.AuthorName, reply.AuthorUsername) + ": " + reply.Content
	}
	return context
}

func authorLabel(name, username string) string {
	if username != "" {
		return "@" + username
	}
	if name != "" {
		return name
	}
	return "Usuário"
}

func (s *Service) report(err error) {
	if s.logError != nil {
		s.logError(err)
	}
}
