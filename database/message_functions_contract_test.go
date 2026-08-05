package database

import (
	"regexp"
	"strings"
	"testing"
)

func TestMessageFunctionsMigrationCreateMessageRejectsInvalidInputsAndCreatesEveryLevel(t *testing.T) {
	sql := readMessageFunctionsMigration(t)
	body := messageFunctionBody(t, sql, "create_message", "int")

	for _, requirement := range []string{
		`(?is)if\s+p_id_parent\s+is\s+null\s+then\s+raise\s+exception\s+'post requires an id_subject parent'`,
		`(?is)insert\s+into\s+posts\s*\(\s*id_user\s*,\s*id_subject\s*,\s*title\s*,\s*image_url\s*,\s*image_description\s*,\s*content\s*\)`,
		`(?is)insert\s+into\s+comments\s*\(\s*id_post\s*,\s*id_user\s*,\s*content\s*\)`,
		`(?is)insert\s+into\s+comments_on_comments\s*\(\s*id_comment\s*,\s*id_user\s*,\s*content\s*\)`,
		`(?is)else\s+raise\s+exception\s+'invalid message type: %'\s*,\s*p_message_type`,
		`(?is)return\s+v_id`,
	} {
		assertMessageContract(t, body, requirement)
	}
	if got := strings.Count(body, "IF p_id_parent IS NULL THEN"); got != 3 {
		t.Fatalf("parent validation count = %d, want 3", got)
	}
}

func TestMessageFunctionsMigrationReactionTogglesEveryTargetAndEnforcesUniqueness(t *testing.T) {
	sql := readMessageFunctionsMigration(t)
	body := messageFunctionBody(t, sql, "reaction", "void")
	assertMessageContract(t, body, `(?is)p_reaction_type\s+not\s+in\s*\(\s*'like'\s*,\s*'dislike'\s*\)\s+then\s+raise\s+exception`)
	assertMessageContract(t, body, `(?is)else\s+raise\s+exception\s+'invalid message type: %'\s*,\s*p_message_type`)

	for _, column := range []string{"id_post", "id_comment", "id_comment_on_comment"} {
		assertMessageContract(t, body, `(?is)select\s+reaction_type\s+into\s+v_reaction\s+from\s+reactions\s+where\s+id_user\s*=\s*p_id_user\s+and\s+`+column+`\s*=\s*p_id_message`)
		assertMessageContract(t, body, `(?is)insert\s+into\s+reactions\s*\(\s*id_user\s*,\s*`+column+`\s*,\s*reaction_type\s*\)`)
		assertMessageContract(t, body, `(?is)delete\s+from\s+reactions\s+where\s+id_user\s*=\s*p_id_user\s+and\s+`+column+`\s*=\s*p_id_message`)
		assertMessageContract(t, body, `(?is)update\s+reactions\s+set\s+reaction_type\s*=\s*p_reaction_type\s+where\s+id_user\s*=\s*p_id_user\s+and\s+`+column+`\s*=\s*p_id_message`)
	}

	indexes := regexp.MustCompile(`(?is)create\s+unique\s+index\s+if\s+not\s+exists\s+\w+\s+on\s+reactions\s*\(\s*id_user\s*,\s*(id_post|id_comment|id_comment_on_comment)\s*\)\s+where\s+(?:id_post|id_comment|id_comment_on_comment)\s+is\s+not\s+null\s*;`).FindAllStringSubmatch(sql, -1)
	if got := len(indexes); got != 3 {
		t.Fatalf("partial reaction unique index count = %d, want 3", got)
	}
	seen := map[string]bool{}
	for _, index := range indexes {
		seen[index[1]] = true
	}
	for _, column := range []string{"id_post", "id_comment", "id_comment_on_comment"} {
		if !seen[column] {
			t.Errorf("missing partial unique reaction index for %s", column)
		}
	}
}

func TestMessageFunctionsMigrationReactionTriggerRecountsEveryTargetOnEveryMutation(t *testing.T) {
	sql := readMessageFunctionsMigration(t)
	body := messageFunctionBody(t, sql, "update_reaction_count", "trigger")
	assertMessageContract(t, body, `(?is)if\s+tg_op\s*=\s*'delete'\s+then`)
	assertMessageContract(t, body, `(?is)else\s+v_post_id\s*:=\s*new\.id_post`)
	for _, target := range []struct{ table, column string }{
		{"posts", "id_post"},
		{"comments", "id_comment"},
		{"comments_on_comments", "id_comment_on_comment"},
	} {
		assertMessageContract(t, body, `(?is)update\s+`+target.table+`\s+set\s+likes\s*=\s*greatest\s*\(\s*0\s*,\s*\(\s*select\s+count\(\*\)\s+from\s+reactions\s+where\s+`+target.column+`\s*=.*?reaction_type\s*=\s*'like'\s*\)\s*\)`)
		assertMessageContract(t, body, `(?is)dislikes\s*=\s*greatest\s*\(\s*0\s*,\s*\(\s*select\s+count\(\*\)\s+from\s+reactions\s+where\s+`+target.column+`\s*=.*?reaction_type\s*=\s*'dislike'\s*\)\s*\)`)
	}
	assertMessageContract(t, sql, `(?is)create\s+trigger\s+trg_update_reaction_count\s+after\s+insert\s+or\s+update\s+or\s+delete\s+on\s+reactions`)
}

func TestMessageFunctionsMigrationMentionsOnlyExistingOtherUsersAndResolvePost(t *testing.T) {
	sql := readMessageFunctionsMigration(t)
	body := messageFunctionBody(t, sql, "notify_mentions", "trigger")
	for _, requirement := range []string{
		`(?is)regexp_matches\s*\(\s*new\.content\s*,\s*'@\(\[a-zA-Z0-9_\]\+\)'\s*,\s*'g'\s*\)`,
		`(?is)select\s+id\s+into\s+v_id_user_mentioned\s+from\s+users\s+where\s+username\s*=\s*v_username`,
		`(?is)v_id_user_mentioned\s+is\s+not\s+null\s+and\s+v_id_user_mentioned\s*<>\s*new\.id_user`,
		`(?is)from\s+comments\s+join\s+posts\s+on\s+posts\.id\s*=\s+comments\.id_post\s+where\s+comments\.id\s*=\s+new\.id_comment`,
		`(?is)insert\s+into\s+notifications\s*\(\s*id_post\s*,\s*id_user\s*,\s*text_show\s*\)`,
	} {
		assertMessageContract(t, body, requirement)
	}
	triggers := regexp.MustCompile(`(?is)create\s+trigger\s+trg_notify_mentions_\w+\s+after\s+insert\s+on\s+(posts|comments|comments_on_comments)\s+for\s+each\s+row\s+execute\s+function\s+notify_mentions\s*\(\s*\)\s*;`).FindAllStringSubmatch(sql, -1)
	if got := len(triggers); got != 3 {
		t.Fatalf("mention trigger count = %d, want 3", got)
	}
	seen := map[string]bool{}
	for _, trigger := range triggers {
		seen[trigger[1]] = true
	}
	for _, table := range []string{"posts", "comments", "comments_on_comments"} {
		if !seen[table] {
			t.Errorf("missing INSERT mention trigger for %s", table)
		}
	}
	if strings.Contains(body, "TG_TABLE_NAME = 'reactions'") {
		t.Fatal("notification function must not process reactions")
	}
}

func messageFunctionBody(t *testing.T, sql, name, returns string) string {
	t.Helper()
	pattern := regexp.MustCompile(`(?is)create\s+or\s+replace\s+function\s+` + regexp.QuoteMeta(name) + `\s*\([^$]*?\)\s*returns\s+` + returns + `\s+language\s+plpgsql\s+as\s+\$\$(.*?)\$\$\s*;`)
	match := pattern.FindStringSubmatch(sql)
	if match == nil {
		t.Fatalf("function body for %s not found", name)
	}
	return match[1]
}

func assertMessageContract(t *testing.T, sql, pattern string) {
	t.Helper()
	if !regexp.MustCompile(pattern).MatchString(sql) {
		t.Errorf("missing required SQL contract %q", pattern)
	}
}
