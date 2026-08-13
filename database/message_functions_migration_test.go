package database

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestMessageFunctionsMigrationContainsFeatureContracts(t *testing.T) {
	sql := readMigrationSQL(t, "migrations/0005_message_functions_triggers.sql")
	for _, pattern := range []string{
		`(?is)create\s+or\s+replace\s+function\s+create_message\s*\(.*p_id_user\s+bigint.*p_id_parent\s+bigint.*\)\s*returns\s+bigint`,
		`(?is)create\s+or\s+replace\s+function\s+reaction\s*\(.*p_id_user\s+bigint.*p_id_message\s+bigint.*\)\s*returns\s+void`,
		`(?is)create\s+or\s+replace\s+function\s+mark_notifications_as_read\s*\(\s*p_id_user\s+bigint\s*\)\s*returns\s+void`,
		`(?is)create\s+or\s+replace\s+function\s+update_reaction_count`,
		`(?is)create\s+or\s+replace\s+function\s+notify_mentions`,
		`(?is)create\s+unique\s+index\s+if\s+not\s+exists.*id_user\s*,\s*id_post`,
		`(?is)create\s+unique\s+index\s+if\s+not\s+exists.*id_user\s*,\s*id_comment`,
		`(?is)create\s+unique\s+index\s+if\s+not\s+exists.*id_user\s*,\s*id_comment_on_comment`,
		`(?is)create\s+trigger\s+trg_update_reaction_count\s+after\s+insert\s+or\s+update\s+or\s+delete\s+on\s+reactions`,
		`(?is)create\s+trigger\s+trg_notify_mentions_\w+\s+after\s+insert\s+on\s+posts`,
		`(?is)create\s+trigger\s+trg_notify_mentions_\w+\s+after\s+insert\s+on\s+comments`,
		`(?is)create\s+trigger\s+trg_notify_mentions_\w+\s+after\s+insert\s+on\s+comments_on_comments`,
	} {
		if !regexp.MustCompile(pattern).MatchString(sql) {
			t.Errorf("migration missing SQL contract %q", pattern)
		}
	}
	if strings.Contains(strings.ToLower(sql), "tg_table_name = 'reactions'") {
		t.Fatal("mention trigger must not process reactions")
	}
}

func TestAcademicSeedMigrationAllowsAdminWithoutAcademicYear(t *testing.T) {
	sql := readMigrationSQL(t, "migrations/0004_admin_year_and_academic_seed.sql")
	for _, required := range []string{
		`(?is)alter\s+table\s+users\s+alter\s+column\s+id_year\s+drop\s+not\s+null`,
		`(?is)update\s+users\s+set\s+id_year\s*=\s*null\s+where\s+is_admin\s*=\s*true`,
		`(?is)users_admin_year_check`,
	} {
		if !regexp.MustCompile(required).MatchString(sql) {
			t.Errorf("academic migration missing SQL contract %q", required)
		}
	}
}

func readMigrationSQL(t *testing.T, path string) string {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read migration %s: %v", path, err)
	}
	return strings.ReplaceAll(string(contents), "\r\n", "\n")
}
