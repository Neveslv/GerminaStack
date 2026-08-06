package database

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestMessageFunctionsMigrationDefinesCreateMessageContract(t *testing.T) {
	sql := readMessageFunctionsMigration(t)
	assertMessageSQL(t, sql, `(?is)create\s+or\s+replace\s+function\s+create_message\s*\(\s*p_message_type\s+text\s*,\s*p_id_user\s+int\s*,\s*p_id_parent\s+int\s*,\s*p_content\s+text\s*,\s*p_title\s+text\s+default\s+null\s*,\s*p_image_url\s+text\s+default\s+null\s*,\s*p_image_description\s+text\s+default\s+null\s*\)\s*returns\s+int`)
	for _, target := range []string{"posts", "comments", "comments_on_comments"} {
		assertMessageSQL(t, sql, `(?is)insert\s+into\s+`+target)
	}
}

func TestMessageFunctionsMigrationDefinesReactionContract(t *testing.T) {
	sql := readMessageFunctionsMigration(t)
	assertMessageSQL(t, sql, `(?is)create\s+or\s+replace\s+function\s+reaction\s*\(\s*p_id_user\s+int\s*,\s*p_id_message\s+int\s*,\s*p_message_type\s+text\s*,\s*p_reaction_type\s+text\s*\)\s*returns\s+void`)
	assertMessageSQL(t, sql, `(?is)p_reaction_type\s+not\s+in\s*\(\s*'like'\s*,\s*'dislike'\s*\)`)
	for _, column := range []string{"id_post", "id_comment", "id_comment_on_comment"} {
		assertMessageSQL(t, sql, `(?is)create\s+unique\s+index\s+if\s+not\s+exists.*?\(\s*id_user\s*,\s*`+column+`\s*\).*?where\s+`+column+`\s+is\s+not\s+null`)
	}
}

func TestMessageFunctionsMigrationDefinesNotificationAndTriggerContracts(t *testing.T) {
	sql := readMessageFunctionsMigration(t)
	assertMessageSQL(t, sql, `(?is)create\s+or\s+replace\s+function\s+mark_notifications_as_read\s*\(\s*p_id_user\s+int\s*\)\s*returns\s+void`)
	assertMessageSQL(t, sql, `(?is)update\s+notifications\s+set\s+is_read\s*=\s*true\s+where\s+id_user\s*=\s*p_id_user\s+and\s+is_read\s*=\s*false`)
	assertMessageSQL(t, sql, `(?is)create\s+or\s+replace\s+function\s+update_reaction_count`)
	assertMessageSQL(t, sql, `(?is)create\s+trigger\s+trg_update_reaction_count\s+after\s+insert\s+or\s+update\s+or\s+delete\s+on\s+reactions`)
	assertMessageSQL(t, sql, `(?is)create\s+or\s+replace\s+function\s+notify_mentions`)
	for _, table := range []string{"posts", "comments", "comments_on_comments"} {
		assertMessageSQL(t, sql, `(?is)create\s+trigger\s+trg_notify_mentions_\w+\s+after\s+insert\s+on\s+`+table)
	}
	if regexp.MustCompile(`(?is)tg_table_name\s*=\s*'reactions'`).MatchString(sql) {
		t.Fatal("mentions must not be generated for reactions")
	}
}

func TestMessageFunctionsMigrationIsIdempotent(t *testing.T) {
	sql := readMessageFunctionsMigration(t)
	for _, function := range []string{"create_message", "reaction", "mark_notifications_as_read", "update_reaction_count", "notify_mentions"} {
		assertMessageSQL(t, sql, `(?is)create\s+or\s+replace\s+function\s+`+function)
	}
	for _, trigger := range []struct{ name, table string }{
		{"trg_update_reaction_count", "reactions"},
		{"trg_notify_mentions_post", "posts"},
		{"trg_notify_mentions_comment", "comments"},
		{"trg_notify_mentions_comment_on_comment", "comments_on_comments"},
	} {
		assertMessageSQL(t, sql, `(?is)drop\s+trigger\s+if\s+exists\s+`+trigger.name+`\s+on\s+`+trigger.table)
	}
}

func readMessageFunctionsMigration(t *testing.T) string {
	t.Helper()
	migration, err := os.ReadFile("migrations/0005_message_functions_triggers.sql")
	if err != nil {
		t.Fatalf("read message functions migration: %v", err)
	}
	return strings.ReplaceAll(string(migration), "\r\n", "\n")
}

func assertMessageSQL(t *testing.T, sql, pattern string) {
	t.Helper()
	if !regexp.MustCompile(pattern).MatchString(sql) {
		t.Errorf("SQL does not match required pattern %q", pattern)
	}
}
