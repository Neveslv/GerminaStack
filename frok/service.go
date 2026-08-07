package frok

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"germinaStack/model"
)

var mentionPattern = regexp.MustCompile(`(?i)(^|[^[:alnum:]_])@frok($|[^[:alnum:]_])`)

type Repository interface {
	BotUserID(context.Context) (int64, error)
	Username(context.Context, int64) (string, error)
	CreateComment(context.Context, int64, int64, string) error
	CreateReply(context.Context, int64, int64, string) error
}

type ReplyClient interface {
	Reply(context.Context, string) (string, error)
}

type Service struct {
	repository Repository
	client     ReplyClient
	timeout    time.Duration
	logError   func(error)
}

func NewService(repository Repository, client ReplyClient, timeout time.Duration, logError func(error)) *Service {
	return &Service{repository: repository, client: client, timeout: timeout, logError: logError}
}

func IsMentioned(content string) bool {
	return mentionPattern.MatchString(content)
}

func (s *Service) DispatchPost(authorID int64, post model.Post) {
	if !IsMentioned(post.Title + "\n" + post.Content) {
		return
	}
	go s.respond(authorID, post.ID, 0, postContext(post))
}

func (s *Service) DispatchComment(authorID int64, comment model.Comment) {
	if !IsMentioned(comment.Content) {
		return
	}
	go s.respond(authorID, 0, comment.ID, "Comentário: "+comment.Content)
}

func (s *Service) respond(authorID, postID, commentID int64, input string) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	username, err := s.repository.Username(ctx, authorID)
	if err != nil {
		s.report(err)
		return
	}
	reply, err := s.client.Reply(ctx, input)
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
	}
}

func formatReply(username, reply string) string {
	reply = strings.TrimSpace(strings.ReplaceAll(reply, "@", "＠"))
	return "@" + username + " " + reply
}

func postContext(post model.Post) string {
	context := fmt.Sprintf("Título: %s\n\nPost: %s", post.Title, post.Content)
	if post.ImageDescription != nil && strings.TrimSpace(*post.ImageDescription) != "" {
		context += "\n\nDescrição da imagem (alt): " + strings.TrimSpace(*post.ImageDescription)
	}
	return context
}

func (s *Service) report(err error) {
	if s.logError != nil {
		s.logError(err)
	}
}
