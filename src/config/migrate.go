package config

import (
	"fmt"
	"log"
)

// migrationColumn describes a single ADD-COLUMN migration step.
type migrationColumn struct {
	table    string
	column   string
	alterSQL string
}

// userSignMigrations lists every new column that must be present on the
// user_sign table.  Each entry is applied exactly once: the column is added
// only when it does not already exist in the live schema.
var userSignMigrations = []migrationColumn{
	{
		table:  "user_sign",
		column: "notify_mode",
		alterSQL: "ALTER TABLE user_sign ADD COLUMN notify_mode INT NOT NULL DEFAULT 0" +
			" COMMENT '签到通知模式 0-全部通知 1-仅失败通知 2-仅成功通知'",
	},
	{
		table:  "user_sign",
		column: "ap_remind",
		alterSQL: "ALTER TABLE user_sign ADD COLUMN ap_remind INT NOT NULL DEFAULT 0" +
			" COMMENT '理智提醒开关 0-关闭 1-开启'",
	},
	{
		table:  "user_sign",
		column: "ap_threshold",
		alterSQL: "ALTER TABLE user_sign ADD COLUMN ap_threshold INT NOT NULL DEFAULT 80" +
			" COMMENT '理智提醒阈值百分比'",
	},
	{
		table:  "user_sign",
		column: "ap_notified",
		alterSQL: "ALTER TABLE user_sign ADD COLUMN ap_notified INT NOT NULL DEFAULT 0" +
			" COMMENT '理智提醒是否已通知 0-未通知 1-已通知'",
	},
}

// MigrateDB applies all pending schema migrations to the connected database.
// Each migration step is idempotent: a column is only added when it is not
// already present in the live schema.
func MigrateDB() error {
	if DBEngine == nil {
		return fmt.Errorf("migrate: database engine is not initialized")
	}
	for _, m := range userSignMigrations {
		if err := applyColumnIfMissing(m); err != nil {
			return err
		}
	}
	return nil
}

// applyColumnIfMissing adds m.column to m.table when it does not already exist.
// It queries INFORMATION_SCHEMA.COLUMNS so the check works on any MySQL-compatible
// database without relying on database-specific DDL extensions.
func applyColumnIfMissing(m migrationColumn) error {
	var count int64
	res := DBEngine.Raw(
		`SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
		 WHERE TABLE_SCHEMA = DATABASE()
		   AND TABLE_NAME   = ?
		   AND COLUMN_NAME  = ?`,
		m.table, m.column,
	).Scan(&count)
	if res.Error != nil {
		return fmt.Errorf("migrate: checking column %s.%s: %w", m.table, m.column, res.Error)
	}
	if count > 0 {
		return nil // already present
	}
	if err := DBEngine.Exec(m.alterSQL).Error; err != nil {
		return fmt.Errorf("migrate: adding column %s.%s: %w", m.table, m.column, err)
	}
	log.Printf("migrate: added column %s.%s", m.table, m.column)
	return nil
}
