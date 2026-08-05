package database

import "testing"

func TestMessageFunctionsMigrationReactionDeleteUsesOldTargetForEveryLevel(t *testing.T) {
	body := messageFunctionBody(t, readMessageFunctionsMigration(t), "update_reaction_count", "trigger")
	for _, target := range []struct{ column, variable string }{
		{"id_post", "v_post_id"},
		{"id_comment", "v_comment_id"},
		{"id_comment_on_comment", "v_reply_id"},
	} {
		assertMessageContract(t, body, `(?is)if\s+tg_op\s*=\s*'delete'\s+then.*?`+target.variable+`\s*:=\s*old\.`+target.column)
	}
}

func TestMessageFunctionsMigrationMentionResolvesPostForEveryMessageLevel(t *testing.T) {
	body := messageFunctionBody(t, readMessageFunctionsMigration(t), "notify_mentions", "trigger")
	for _, requirement := range []string{
		`(?is)if\s+tg_table_name\s*=\s*'posts'\s+then\s+v_post_id\s*:=\s*new\.id`,
		`(?is)elsif\s+tg_table_name\s*=\s*'comments'\s+then\s+v_post_id\s*:=\s*new\.id_post`,
		`(?is)elsif\s+tg_table_name\s*=\s*'comments_on_comments'\s+then.*?from\s+comments\s+join\s+posts`,
	} {
		assertMessageContract(t, body, requirement)
	}
}