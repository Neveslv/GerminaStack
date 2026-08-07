package frok

import (
	"context"
	"testing"
	"time"

	"germinaStack/model"
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

func TestPostContextIncludesImageDescription(t *testing.T) {
	t.Parallel()
	description := "Diagrama com as tabelas e chaves estrangeiras."
	context := postContext(model.Post{Title: "Dúvida", Content: "@frok explica", ImageDescription: &description})
	if context != "Título: Dúvida\n\nPost: @frok explica\n\nDescrição da imagem (alt): Diagrama com as tabelas e chaves estrangeiras." {
		t.Fatalf("postContext() = %q", context)
	}
}

func TestThreadContextIncludesRelatedMessages(t *testing.T) {
	t.Parallel()
	description := "Fluxo entre cliente e banco."
	context := threadContext(
		model.Post{Title: "Transação", Content: "@frok explica", ImageDescription: &description},
		model.Comment{Content: "Minha dúvida é sobre commit.", AuthorUsername: "ana"},
		[]model.CommentOnComment{
			{Content: "O rollback desfaz.", AuthorUsername: "bruno"},
			{Content: "@frok e o isolamento?", AuthorUsername: "carla"},
		},
	)
	want := "Título: Transação\n\nPost: @frok explica\n\nDescrição da imagem (alt): Fluxo entre cliente e banco.\n\nComentário relacionado de @ana: Minha dúvida é sobre commit.\n\nRespostas relacionadas:\n- @bruno: O rollback desfaz.\n- @carla: @frok e o isolamento?"
	if context != want {
		t.Fatalf("threadContext() = %q", context)
	}
}

func TestServiceRepliesToMentionedThreadWithItsContext(t *testing.T) {
	t.Parallel()
	repository := &repositoryFake{
		username: "ana",
		botID:    7,
		post:     model.Post{ID: 8, Title: "Transação", Content: "Texto do post"},
		comment:  model.Comment{ID: 9, PostID: 8, Content: "@frok explique", AuthorUsername: "ana"},
		replies:  []model.CommentOnComment{{Content: "Primeira resposta", AuthorUsername: "bruno"}},
	}
	client := &recordingClient{reply: "Resposta do Frok"}
	service := NewService(repository, client, time.Second, nil)
	service.respondToThread(42, 9)
	if repository.replyCommentID != 9 || repository.replyContent != "@ana Resposta do Frok" {
		t.Fatalf("reply = %#v", repository)
	}
	if client.input != "Título: Transação\n\nPost: Texto do post\n\nComentário relacionado de @ana: @frok explique\n\nRespostas relacionadas:\n- @bruno: Primeira resposta" {
		t.Fatalf("Frok input = %q", client.input)
	}
}

type clientFake struct{ reply string }

func (f clientFake) Reply(context.Context, string) (string, error) { return f.reply, nil }

type recordingClient struct {
	reply string
	input string
}

func (f *recordingClient) Reply(_ context.Context, input string) (string, error) {
	f.input = input
	return f.reply, nil
}

type repositoryFake struct {
	username       string
	botID          int64
	post           model.Post
	comment        model.Comment
	replies        []model.CommentOnComment
	commentUserID  int64
	commentPostID  int64
	commentContent string
	replyUserID    int64
	replyCommentID int64
	replyContent   string
}

func (f *repositoryFake) BotUserID(context.Context) (int64, error)        { return f.botID, nil }
func (f *repositoryFake) Username(context.Context, int64) (string, error) { return f.username, nil }
func (f *repositoryFake) GetPost(context.Context, int64) (model.Post, error) {
	return f.post, nil
}
func (f *repositoryFake) GetComment(context.Context, int64) (model.Comment, error) {
	return f.comment, nil
}
func (f *repositoryFake) ListReplies(context.Context, int64) ([]model.CommentOnComment, error) {
	return f.replies, nil
}
func (f *repositoryFake) CreateComment(_ context.Context, userID, postID int64, content string) error {
	f.commentUserID, f.commentPostID, f.commentContent = userID, postID, content
	return nil
}
func (f *repositoryFake) CreateReply(_ context.Context, userID, commentID int64, content string) error {
	f.replyUserID, f.replyCommentID, f.replyContent = userID, commentID, content
	return nil
}
