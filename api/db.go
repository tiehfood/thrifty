package main

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

var defaultIconId = "00000000-0000-0000-0000-000000000000"
var defaultUserId = "00000000-0000-0000-0000-000000000001"
var errorPrefix = "Error: "

var migrations = []Migration{
	{
		Version: 1,
		Statements: []string{
			`CREATE TABLE IF NOT EXISTS icons (
				id TEXT NOT NULL PRIMARY KEY,
				data TEXT,
				hash TEXT)`,
			`CREATE TABLE IF NOT EXISTS flows (
				id TEXT NOT NULL PRIMARY KEY,
				name TEXT,
				description TEXT,
				amount FLOAT,
				iconId TEXT DEFAULT '00000000-0000-0000-0000-000000000000',
				FOREIGN KEY (iconId) REFERENCES icons(id))`,
		},
	},
	{
		Version: 2,
		Statements: []string{
			`CREATE TABLE IF NOT EXISTS users (
				id TEXT NOT NULL PRIMARY KEY,
				name TEXT NOT NULL)`,
			`INSERT OR IGNORE INTO users (id, name) VALUES ('00000000-0000-0000-0000-000000000001', 'thrifty')`,
			`CREATE TABLE IF NOT EXISTS settings (
				id INTEGER NOT NULL PRIMARY KEY CHECK (id = 1),
				currentUserId TEXT NOT NULL,
				multiUserEnabled INTEGER NOT NULL DEFAULT 0,
				FOREIGN KEY (currentUserId) REFERENCES users(id))`,
			`INSERT OR IGNORE INTO settings (id, currentUserId, multiUserEnabled) VALUES (1, '00000000-0000-0000-0000-000000000001', 0)`,
			`ALTER TABLE flows ADD COLUMN userId TEXT REFERENCES users(id) DEFAULT '00000000-0000-0000-0000-000000000001'`,
		},
	},
	{
		Version: 3,
		Statements: []string{
			`ALTER TABLE settings ADD COLUMN numberFormat TEXT NOT NULL DEFAULT 'eu-decimal'`,
		},
	},
	{
		Version: 4,
		Statements: []string{
			`CREATE TABLE IF NOT EXISTS groups (
				id TEXT NOT NULL PRIMARY KEY,
				name TEXT,
				description TEXT,
				iconId TEXT DEFAULT '00000000-0000-0000-0000-000000000000',
				userId TEXT REFERENCES users(id),
				FOREIGN KEY (iconId) REFERENCES icons(id))`,
			`ALTER TABLE flows ADD COLUMN groupId TEXT REFERENCES groups(id)`,
		},
	},
}

func runMigrations(dbCon *sql.DB) error {
	_, err := dbCon.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version    INTEGER NOT NULL PRIMARY KEY,
		applied_at TEXT    NOT NULL)`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	rows, err := dbCon.Query(`SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		return fmt.Errorf("failed to query applied migrations: %w", err)
	}
	applied := map[int]bool{}
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			rows.Close()
			return err
		}
		applied[v] = true
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	for _, m := range migrations {
		if applied[m.Version] {
			fmt.Printf("Migration %d already applied, skipping\n", m.Version)
			continue
		}
		fmt.Printf("Applying migration %d...\n", m.Version)
		for _, stmt := range m.Statements {
			if _, err := dbCon.Exec(stmt); err != nil {
				return fmt.Errorf("migration %d failed: %w\nSQL: %s", m.Version, err, stmt)
			}
		}
		_, err = dbCon.Exec(
			`INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)`,
			m.Version, time.Now().UTC().Format(time.RFC3339),
		)
		if err != nil {
			return fmt.Errorf("failed to record migration %d: %w", m.Version, err)
		}
		fmt.Printf("Migration %d applied successfully\n", m.Version)
	}
	return nil
}

func initAndOpenDb() (*sql.DB, error) {
	dataSourceName := os.Getenv("SQLITE_DB_PATH")
	defaultIconFlow := Flow{
		Icon: "data:image/svg+xml;base64,PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0iVVRGLTgiPz4KPHN2ZyBpZD0iTGF5ZXJfMSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIiB2ZXJzaW9uPSIxLjEiIHZpZXdCb3g9IjAgMCAyNTQuMiAyNTQuMiI+CiAgPGRlZnM+CiAgICA8c3R5bGU+CiAgICAgIC5zdDAgewogICAgICAgIGZpbGw6ICM0NDliMzg7CiAgICAgIH0KCiAgICAgIC5zdDEgewogICAgICAgIGZpbGw6ICNmZmMzMzY7CiAgICAgIH0KCiAgICAgIC5zdDIgewogICAgICAgIGZpbGw6ICNjODkzMTg7CiAgICAgIH0KCiAgICAgIC5zdDMgewogICAgICAgIGZpbGw6ICM2MmM3NTE7CiAgICAgIH0KCiAgICAgIC5zdDQgewogICAgICAgIGZpbGw6ICNlYWIwMjg7CiAgICAgIH0KICAgIDwvc3R5bGU+CiAgPC9kZWZzPgogIDxyZWN0IGNsYXNzPSJzdDAiIHk9IjU0LjYiIHdpZHRoPSIyNTQuMiIgaGVpZ2h0PSIxMzAuNyIvPgogIDxwYXRoIGNsYXNzPSJzdDMiIGQ9Ik0yMTMuNSw3My41SDQwLjhjMCwxMi4xLTkuOCwyMS45LTIxLjksMjEuOXY0OS4yYzEyLjEsMCwyMS45LDkuOCwyMS45LDIxLjloMTcyLjdjMC0uMiwwLS40LDAtLjcsMC0xMi4xLDkuOC0yMS45LDIxLjktMjEuOXYtNDkuMmMtMTEuOCwwLTIxLjUtOS40LTIxLjgtMjEuMloiLz4KICA8Y2lyY2xlIGNsYXNzPSJzdDAiIGN4PSIxMjcuMSIgY3k9IjEyMCIgcj0iMzguOSIvPgogIDxjaXJjbGUgY2xhc3M9InN0MCIgY3g9IjY3IiBjeT0iMTIwIiByPSI3LjgiLz4KICA8Y2lyY2xlIGNsYXNzPSJzdDAiIGN4PSIxODcuMiIgY3k9IjEyMCIgcj0iNy44IiB0cmFuc2Zvcm09InRyYW5zbGF0ZSgzOC44IDI4NS41KSByb3RhdGUoLTgwLjgpIi8+CiAgPHBhdGggY2xhc3M9InN0MyIgZD0iTTEyNi44LDE0MC40Yy01LjksMC0xMS4yLTMuMS0xMy4zLThsNy43LTMuNWMuOCwxLjgsMy4xLDMsNS42LDNzNC44LTEuMiw1LjYtM2MuMy0uNi4zLTEuMi4yLTEuNS0uMi0uNS0xLjQtMS45LTYuNy0zLjUtNC40LTEuMy05LjctMy40LTExLjktNy45LS45LTEuOC0xLjUtNC43LDAtOC41LDIuMi01LDcuNS04LjEsMTMuNC04LjFzMTEuMiwzLjEsMTMuMyw4bC03LjcsMy41Yy0uOC0xLjgtMy4xLTMtNS42LTNzLTQuOCwxLjItNS42LDNjLS4zLjYtLjMsMS4yLS4yLDEuNS4yLjUsMS40LDEuOSw2LjcsMy41LDQuNCwxLjMsOS43LDMuNCwxMS45LDcuOS45LDEuOCwxLjUsNC43LDAsOC41LTIuMiw1LTcuNSw4LjEtMTMuNCw4LjFaIi8+CiAgPHJlY3QgY2xhc3M9InN0MyIgeD0iMTIyLjYiIHk9IjEzNi4yIiB3aWR0aD0iOC41IiBoZWlnaHQ9IjkuOCIvPgogIDxyZWN0IGNsYXNzPSJzdDMiIHg9IjEyMi42IiB5PSI5My45IiB3aWR0aD0iOC41IiBoZWlnaHQ9IjkuOCIvPgogIDxjaXJjbGUgY2xhc3M9InN0MyIgY3g9IjI0LjQiIGN5PSI3OC44IiByPSI1LjUiLz4KICA8Y2lyY2xlIGNsYXNzPSJzdDMiIGN4PSIyNC40IiBjeT0iMTYxIiByPSI1LjUiLz4KICA8Y2lyY2xlIGNsYXNzPSJzdDMiIGN4PSIyMjkuOCIgY3k9Ijc4LjkiIHI9IjUuNSIvPgogIDxjaXJjbGUgY2xhc3M9InN0MyIgY3g9IjIyOS44IiBjeT0iMTYxLjEiIHI9IjUuNSIvPgogIDxjaXJjbGUgY2xhc3M9InN0NCIgY3g9IjIxNy43IiBjeT0iMTY2LjYiIHI9IjI5LjUiLz4KICA8Y2lyY2xlIGNsYXNzPSJzdDEiIGN4PSIyMTcuNyIgY3k9IjE2Ni42IiByPSIyMy45Ii8+CiAgPHBhdGggY2xhc3M9InN0MiIgZD0iTTIxNy41LDE3OS4xYy0zLjYsMC02LjktMS45LTguMi00LjlsNC44LTIuMWMuNSwxLjEsMS45LDEuOCwzLjQsMS44czMtLjgsMy40LTEuOGMuMi0uNC4yLS43LjEtLjktLjEtLjMtLjktMS4xLTQuMS0yLjEtMi43LS44LTYtMi4xLTcuMy00LjktLjUtMS4xLS45LTIuOSwwLTUuMiwxLjQtMy4xLDQuNi01LDguMi01czYuOSwxLjksOC4yLDQuOWwtNC44LDIuMWMtLjUtMS4xLTEuOS0xLjgtMy40LTEuOHMtMywuOC0zLjQsMS44Yy0uMi40LS4yLjctLjEuOS4xLjMuOSwxLjEsNC4xLDIuMSwyLjcuOCw2LDIuMSw3LjMsNC45LjUsMS4xLjksMi45LDAsNS4yLTEuNCwzLjEtNC42LDUtOC4yLDVaIi8+CiAgPHJlY3QgY2xhc3M9InN0MiIgeD0iMjE0LjkiIHk9IjE3Ni41IiB3aWR0aD0iNS4yIiBoZWlnaHQ9IjYiLz4KICA8cmVjdCBjbGFzcz0ic3QyIiB4PSIyMTQuOSIgeT0iMTUwLjYiIHdpZHRoPSI1LjIiIGhlaWdodD0iNiIvPgogIDxjaXJjbGUgY2xhc3M9InN0NCIgY3g9IjE3MC42IiBjeT0iMTgzLjEiIHI9IjI5LjUiIHRyYW5zZm9ybT0idHJhbnNsYXRlKC0zNy41IDMyMi4xKSByb3RhdGUoLTgwLjgpIi8+CiAgPGNpcmNsZSBjbGFzcz0ic3QxIiBjeD0iMTcwLjUiIGN5PSIxODMuMSIgcj0iMjMuOSIvPgogIDxwYXRoIGNsYXNzPSJzdDIiIGQ9Ik0xNzAuMywxOTUuN2MtMy42LDAtNi45LTEuOS04LjItNC45bDQuOC0yLjFjLjUsMS4xLDEuOSwxLjgsMy40LDEuOHMzLS44LDMuNC0xLjhjLjItLjQuMi0uNy4xLS45LS4xLS4zLS45LTEuMS00LjEtMi4xLTIuNy0uOC02LTIuMS03LjMtNC45LS41LTEuMS0uOS0yLjksMC01LjIsMS40LTMuMSw0LjYtNSw4LjItNXM2LjksMS45LDguMiw0LjlsLTQuOCwyLjFjLS41LTEuMS0xLjktMS44LTMuNC0xLjhzLTMsLjgtMy40LDEuOGMtLjIuNC0uMi43LS4xLjkuMS4zLjksMS4xLDQuMSwyLjEsMi43LjgsNiwyLjEsNy4zLDQuOS41LDEuMS45LDIuOSwwLDUuMi0xLjQsMy4xLTQuNiw1LTguMiw1WiIvPgogIDxyZWN0IGNsYXNzPSJzdDIiIHg9IjE2Ny43IiB5PSIxOTMuMSIgd2lkdGg9IjUuMiIgaGVpZ2h0PSI2Ii8+CiAgPHJlY3QgY2xhc3M9InN0MiIgeD0iMTY3LjciIHk9IjE2Ny4xIiB3aWR0aD0iNS4yIiBoZWlnaHQ9IjYiLz4KICA8Y2lyY2xlIGNsYXNzPSJzdDQiIGN4PSIyMTQuOCIgY3k9IjIwNS42IiByPSIyOS41Ii8+CiAgPGNpcmNsZSBjbGFzcz0ic3QxIiBjeD0iMjE0LjgiIGN5PSIyMDUuNiIgcj0iMjMuOSIvPgogIDxwYXRoIGNsYXNzPSJzdDIiIGQ9Ik0yMTQuNiwyMTguMmMtMy42LDAtNi45LTEuOS04LjItNC45bDQuOC0yLjFjLjUsMS4xLDEuOSwxLjgsMy40LDEuOHMzLS44LDMuNC0xLjhjLjItLjQuMi0uNy4xLS45LS4xLS4zLS45LTEuMS00LjEtMi4xLTIuNy0uOC02LTIuMS03LjMtNC45LS41LTEuMS0uOS0yLjksMC01LjIsMS40LTMuMSw0LjYtNSw4LjItNXM2LjksMS45LDguMiw0LjlsLTQuOCwyLjFjLS41LTEuMS0xLjktMS44LTMuNC0xLjhzLTMsLjgtMy40LDEuOGMtLjIuNC0uMi43LS4xLjkuMS4zLjksMS4xLDQuMSwyLjEsMi43LjgsNiwyLjEsNy4zLDQuOS41LDEuMS45LDIuOSwwLDUuMi0xLjQsMy4xLTQuNiw1LTguMiw1WiIvPgogIDxyZWN0IGNsYXNzPSJzdDIiIHg9IjIxMiIgeT0iMjE1LjYiIHdpZHRoPSI1LjIiIGhlaWdodD0iNiIvPgogIDxyZWN0IGNsYXNzPSJzdDIiIHg9IjIxMiIgeT0iMTg5LjYiIHdpZHRoPSI1LjIiIGhlaWdodD0iNiIvPgo8L3N2Zz4=",
	}

	if len(dataSourceName) == 0 {
		dataSourceName = "thrifty.sqlite"
	}
	dbCon, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, err
	}

	printSqlVersion(dbCon)

	if err := runMigrations(dbCon); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	insertIcon(dbCon, defaultIconFlow, &defaultIconId)
	return dbCon, nil
}

func printSqlVersion(dbCon *sql.DB) {
	query := "SELECT sqlite_version();"
	row, err := getSqlRow(dbCon, query, nil)
	if err != nil {
		fmt.Println("Error: ", err)
	}
	fmt.Println("SQLite version: ", row[0])
}

func execSql(dbCon *sql.DB, query string, args ...interface{}) {
	result, err := dbCon.Exec(query, args...)
	if err != nil {
		fmt.Println(errorPrefix, err)
	}

	id, _ := result.LastInsertId()
	numRows, _ := result.RowsAffected()
	fmt.Printf("LastInsertId %d, RowsAffected: %d\n", id, numRows)
}

func getSqlRow(dbCon *sql.DB, query string, args ...interface{}) ([]interface{}, error) {
	rows, err := dbCon.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	row := make([]interface{}, len(columns))
	scanArgs := make([]interface{}, len(columns))
	for i := range row {
		scanArgs[i] = &row[i]
	}

	if rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}
	}
	return row, rows.Err()
}

func getMD5(input string) string {
	hash := md5.Sum([]byte(input))
	return fmt.Sprintf("%x", hash)
}

func getDbConnection(c *gin.Context) *sql.DB {
	dbConVar, exists := c.Get("dbCon")
	if !exists {
		c.JSON(500, HTTPError{Error: "Database connection not found"})
		return nil
	}
	dbCon, ok := dbConVar.(*sql.DB)
	if !ok {
		c.JSON(500, HTTPError{Error: "Invalid database connection"})
		return nil
	}
	return dbCon
}

func isIconUnused(dbCon *sql.DB, iconId string) bool {
	var total int
	err := dbCon.QueryRow(`
		SELECT (SELECT count(*) FROM flows WHERE iconId = ?) +
		       (SELECT count(*) FROM groups WHERE iconId = ?)`,
		iconId, iconId).Scan(&total)
	return err == nil && total == 0
}

func getCurrentUserId(dbCon *sql.DB) (string, error) {
	var currentUserId string
	err := dbCon.QueryRow("SELECT currentUserId FROM settings WHERE id = 1").Scan(&currentUserId)
	if err != nil {
		return "", fmt.Errorf("current user not found: %w", err)
	}
	return currentUserId, nil
}

func insertIcon(dbCon *sql.DB, flow Flow, iconId *string) string {
	hash := getMD5(flow.Icon)

	if iconId == nil {
		var id = uuid.New().String()
		iconId = &id
	}

	row, err := getSqlRow(dbCon, "SELECT id, data, hash FROM icons WHERE hash = ?", hash)
	if err != nil {
		fmt.Println(errorPrefix, err)
	}

	if row[0] == nil {
		execSql(dbCon, "INSERT OR REPLACE INTO icons (id, data, hash) VALUES (?, ?, ?)", *iconId, flow.Icon, hash)
	} else {
		value := row[0].(string)
		iconId = &value
	}
	return *iconId
}
