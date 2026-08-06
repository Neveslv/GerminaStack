package model

import "testing"

func TestPostValidateForCreateRejectsInvalidIDsAndText(t *testing.T) {
	tests := []struct {
		name string
		post Post
	}{
		{
			name: "missing user",
			post: Post{SubjectID: 2, Title: "Title", Content: "Content"},
		},
		{
			name: "missing subject",
			post: Post{UserID: 3, Title: "Title", Content: "Content"},
		},
		{
			name: "blank title",
			post: Post{UserID: 3, SubjectID: 2, Title: "  ", Content: "Content"},
		},
		{
			name: "blank content",
			post: Post{UserID: 3, SubjectID: 2, Title: "Title", Content: "\n"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.post.ValidateForCreate(); err == nil {
				t.Fatal("ValidateForCreate() error = nil, want validation error")
			}
		})
	}
}

func TestPostValidateForCreateAcceptsCompleteMessage(t *testing.T) {
	post := Post{
		UserID:           3,
		SubjectID:        2,
		Title:            "A title",
		Content:          "A body",
		ImageURL:         stringPointer("diagram.png"),
		ImageDescription: stringPointer("A diagram"),
	}

	if err := post.ValidateForCreate(); err != nil {
		t.Fatalf("ValidateForCreate() error = %v, want nil", err)
	}
}

func TestCommentValidateForCreateRejectsInvalidIDsAndContent(t *testing.T) {
	tests := []Comment{
		{UserID: 3, Content: "Content"},
		{PostID: 2, Content: "Content"},
		{UserID: 3, PostID: 2, Content: "  "},
	}

	for index, comment := range tests {
		t.Run(testName(index), func(t *testing.T) {
			if err := comment.ValidateForCreate(); err == nil {
				t.Fatal("ValidateForCreate() error = nil, want validation error")
			}
		})
	}
}

func TestCommentOnCommentValidateForCreateRejectsInvalidIDsAndContent(t *testing.T) {
	tests := []CommentOnComment{
		{UserID: 3, Content: "Content"},
		{CommentID: 2, Content: "Content"},
		{UserID: 3, CommentID: 2, Content: "\t"},
	}

	for index, reply := range tests {
		t.Run(testName(index), func(t *testing.T) {
			if err := reply.ValidateForCreate(); err == nil {
				t.Fatal("ValidateForCreate() error = nil, want validation error")
			}
		})
	}
}

func TestMessageValidationAcceptsCompleteCommentsAndReplies(t *testing.T) {
	if err := (Comment{UserID: 3, PostID: 2, Content: "Comment"}).ValidateForCreate(); err != nil {
		t.Fatalf("Comment.ValidateForCreate() error = %v, want nil", err)
	}
	if err := (CommentOnComment{UserID: 3, CommentID: 2, Content: "Reply"}).ValidateForCreate(); err != nil {
		t.Fatalf("CommentOnComment.ValidateForCreate() error = %v, want nil", err)
	}
}

func testName(index int) string {
	return []string{"missing parent or user", "missing parent or user", "blank content"}[index]
}
