package frok

import (
	"context"
	"testing"
	"time"
)

func TestIsMentionedMatchesOnlyFrok(t *testing.T) {
	t.Parallel()
	for _, text := range []string{"@frok explique", "Oi, @FROK!", "pergunte ao @frok."} {
		if !IsMentioned(text) {
			t.Fatalf("IsMentioned(%q) = false", text)
		}
	}
	for _, text := range []string{"@frokzinho", "email@frok.com", "sem menção"} {
		if IsMentioned(text) {
			t.Fatalf("IsMentioned(%q) = true", text)
		}
	}
}

func TestServiceCreatesBotReplyAndNotifiesOnlyAuthor(t *testing.T) {
	t.Parallel()
	repository := &repositoryFake{username: "ana.silva", botID: 7}
	service := NewService(repository, clientFake{reply: "@outra-pessoa resposta"}, time.Second, nil)
	service.respond(42, 0, 9, "Comentário: @frok ajude")
	if repository.replyCommentID != 9 || repository.replyUserID != 7 {
		t.Fatalf("reply target/user = %d/%d", repository.replyCommentID, repository.replyUserID)
	}
	if repository.replyContent != "@ana.silva ＠outra-pessoa resposta" {
		t.Fatalf("reply content = %q", repository.replyContent)
	}
}

func TestServiceCreatesCommentForMentionedPost(t *testing.T) {
	t.Parallel()
	repository := &repositoryFake{username: "ana", botID: 7}
	service := NewService(repository, clientFake{reply: "Use uma chave primária."}, time.Second, nil)
	service.respond(42, 8, 0, "Post: @frok")
	if repository.commentPostID != 8 || repository.commentUserID != 7 || repository.commentContent != "@ana Use uma chave primária." {
		t.Fatalf("comment = %#v", repository)
	}
}

type clientFake struct{ reply string }

func (f clientFake) Reply(context.Context, string) (string, error) { return f.reply, nil }

type repositoryFake struct {
	username       string
	botID          int64
	commentUserID  int64
	commentPostID  int64
	commentContent string
	replyUserID    int64
	replyCommentID int64
	replyContent   string
}

func (f *repositoryFake) BotUserID(context.Context) (int64, error)        { return f.botID, nil }
func (f *repositoryFake) Username(context.Context, int64) (string, error) { return f.username, nil }
func (f *repositoryFake) CreateComment(_ context.Context, userID, postID int64, content string) error {
	f.commentUserID, f.commentPostID, f.commentContent = userID, postID, content
	return nil
}
func (f *repositoryFake) CreateReply(_ context.Context, userID, commentID int64, content string) error {
	f.replyUserID, f.replyCommentID, f.replyContent = userID, commentID, content
	return nil
}
