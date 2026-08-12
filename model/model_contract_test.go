package model

import (
	"reflect"
	"testing"
	"time"
)

type fieldContract struct {
	name      string
	fieldType reflect.Type
	dbTag     string
	jsonTag   string
}

func contractField(
	name string,
	value any,
	dbTag string,
	jsonTag string,
) fieldContract {
	return fieldContract{
		name:      name,
		fieldType: reflect.TypeOf(value),
		dbTag:     dbTag,
		jsonTag:   jsonTag,
	}
}

func TestModelStructContractsMatchDatabase(t *testing.T) {
	contracts := []struct {
		name   string
		model  any
		fields []fieldContract
	}{
		{
			name:  "Year",
			model: Year{},
			fields: []fieldContract{
				contractField("ID", int64(0), "id", "id"),
				contractField("Year", "", "year", "year"),
				contractField(
					"CreatedAt",
					(*time.Time)(nil),
					"created_at",
					"created_at",
				),
			},
		},
		{
			name:  "Subject",
			model: Subject{},
			fields: []fieldContract{
				contractField("ID", int64(0), "id", "id"),
				contractField("YearID", (*int64)(nil), "id_year", "id_year"),
				contractField("Subject", "", "subject", "subject"),
				contractField(
					"CreatedAt",
					(*time.Time)(nil),
					"created_at",
					"created_at",
				),
				contractField("PostsCount", int64(0), "posts_count", "posts_count"),
			},
		},
		{
			name:  "User",
			model: User{},
			fields: []fieldContract{
				contractField("ID", int64(0), "id", "id"),
				contractField("YearID", int64(0), "id_year", "id_year"),
				contractField("Name", "", "name", "name"),
				contractField(
					"ProfileImageURL",
					(*string)(nil),
					"profile_image_url",
					"profile_image_url",
				),
				contractField(
					"ProfileImageDescription",
					(*string)(nil),
					"profile_image_description",
					"profile_image_description",
				),
				contractField("Username", "", "username", "username"),
				contractField("Email", "", "email", "email"),
				contractField("Password", "", "password", "-"),
				contractField("IsAdmin", false, "is_admin", "is_admin"),
				contractField("IsBanned", false, "is_banned", "is_banned"),
				contractField(
					"CreatedAt",
					(*time.Time)(nil),
					"created_at",
					"created_at",
				),
			},
		},
		{
			name:  "Preference",
			model: Preference{},
			fields: []fieldContract{
				contractField("ID", int64(0), "id", "id"),
				contractField("UserID", int64(0), "id_user", "id_user"),
				contractField(
					"ContrastTheme",
					(*ContrastTheme)(nil),
					"contrast_theme",
					"contrast_theme",
				),
				contractField(
					"FontFamily",
					(*FontFamily)(nil),
					"font_family",
					"font_family",
				),
				contractField(
					"FontSpacing",
					(*FontSpacing)(nil),
					"font_spacing",
					"font_spacing",
				),
				contractField(
					"FontSize",
					(*FontSize)(nil),
					"font_size",
					"font_size",
				),
				contractField(
					"CreatedAt",
					(*time.Time)(nil),
					"created_at",
					"created_at",
				),
			},
		},
		{
			name:  "Post",
			model: Post{},
			fields: []fieldContract{
				contractField("ID", int64(0), "id", "id"),
				contractField("UserID", int64(0), "id_user", "id_user"),
				contractField(
					"SubjectID",
					int64(0),
					"id_subject",
					"id_subject",
				),
				contractField("Title", "", "title", "title"),
				contractField(
					"ImageURL",
					(*string)(nil),
					"image_url",
					"image_url",
				),
				contractField(
					"ImageDescription",
					(*string)(nil),
					"image_description",
					"image_description",
				),
				contractField("Content", "", "content", "content"),
				contractField("Likes", int64(0), "likes", "likes"),
				contractField("Dislikes", int64(0), "dislikes", "dislikes"),
				contractField("CommentsCount", int64(0), "comments_count", "comments_count"),
				contractField("AuthorName", "", "author_name", "author_name"),
				contractField("AuthorUsername", "", "author_username", "author_username"),
				contractField("AuthorImageURL", (*string)(nil), "author_image_url", "author_image_url"),
				contractField("AuthorImageDescription", (*string)(nil), "author_image_description", "author_image_description"),
				contractField(
					"CreatedAt",
					(*time.Time)(nil),
					"created_at",
					"created_at",
				),
			},
		},
		{
			name:  "Comment",
			model: Comment{},
			fields: []fieldContract{
				contractField("ID", int64(0), "id", "id"),
				contractField("PostID", int64(0), "id_post", "id_post"),
				contractField("UserID", int64(0), "id_user", "id_user"),
				contractField("Content", "", "content", "content"),
				contractField("Likes", int64(0), "likes", "likes"),
				contractField("Dislikes", int64(0), "dislikes", "dislikes"),
				contractField("AuthorName", "", "author_name", "author_name"),
				contractField("AuthorUsername", "", "author_username", "author_username"),
				contractField(
					"CreatedAt",
					(*time.Time)(nil),
					"created_at",
					"created_at",
				),
			},
		},
		{
			name:  "CommentOnComment",
			model: CommentOnComment{},
			fields: []fieldContract{
				contractField("ID", int64(0), "id", "id"),
				contractField(
					"CommentID",
					int64(0),
					"id_comment",
					"id_comment",
				),
				contractField("UserID", int64(0), "id_user", "id_user"),
				contractField("Content", "", "content", "content"),
				contractField("Likes", int64(0), "likes", "likes"),
				contractField("Dislikes", int64(0), "dislikes", "dislikes"),
				contractField("AuthorName", "", "author_name", "author_name"),
				contractField("AuthorUsername", "", "author_username", "author_username"),
				contractField(
					"CreatedAt",
					(*time.Time)(nil),
					"created_at",
					"created_at",
				),
			},
		},
		{
			name:  "Reaction",
			model: Reaction{},
			fields: []fieldContract{
				contractField("ID", int64(0), "id", "id"),
				contractField("UserID", int64(0), "id_user", "id_user"),
				contractField("PostID", (*int64)(nil), "id_post", "id_post"),
				contractField(
					"CommentID",
					(*int64)(nil),
					"id_comment",
					"id_comment",
				),
				contractField(
					"CommentOnCommentID",
					(*int64)(nil),
					"id_comment_on_comment",
					"id_comment_on_comment",
				),
				contractField(
					"ReactionType",
					ReactionType(""),
					"reaction_type",
					"reaction_type",
				),
				contractField(
					"CreatedAt",
					(*time.Time)(nil),
					"created_at",
					"created_at",
				),
			},
		},
		{
			name:  "Notification",
			model: Notification{},
			fields: []fieldContract{
				contractField("ID", int64(0), "id", "id"),
				contractField("PostID", (*int64)(nil), "id_post", "id_post"),
				contractField("UserID", int64(0), "id_user", "id_user"),
				contractField("TextShow", "", "text_show", "text_show"),
				contractField("IsRead", false, "is_read", "is_read"),
				contractField("IsHidden", false, "is_hidden", "is_hidden"),
				contractField(
					"CreatedAt",
					(*time.Time)(nil),
					"created_at",
					"created_at",
				),
			},
		},
	}

	for _, contract := range contracts {
		t.Run(contract.name, func(t *testing.T) {
			modelType := reflect.TypeOf(contract.model)
			if got, want := modelType.NumField(), len(contract.fields); got != want {
				t.Fatalf("quantidade de campos = %d, esperado %d", got, want)
			}

			for _, expected := range contract.fields {
				field, found := modelType.FieldByName(expected.name)
				if !found {
					t.Errorf("campo %s não encontrado", expected.name)
					continue
				}
				if field.Type != expected.fieldType {
					t.Errorf(
						"%s type = %s, esperado %s",
						expected.name,
						field.Type,
						expected.fieldType,
					)
				}
				if got := field.Tag.Get("db"); got != expected.dbTag {
					t.Errorf(
						"%s db tag = %q, esperado %q",
						expected.name,
						got,
						expected.dbTag,
					)
				}
				if got := field.Tag.Get("json"); got != expected.jsonTag {
					t.Errorf(
						"%s json tag = %q, esperado %q",
						expected.name,
						got,
						expected.jsonTag,
					)
				}
			}
		})
	}
}
