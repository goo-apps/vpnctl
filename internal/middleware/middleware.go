package middleware

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"github.com/goo-apps/vpnctl/logger"

	_ "github.com/mattn/go-sqlite3"
)

const dbPath = "~/.vpnctl/vpnctl.db"

// expandPath expands ~ to the user home directory.
func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, strings.TrimPrefix(path, "~")), nil
	}
	return path, nil
}

// initDB initializes the SQLite database and creates the helm_charts table if it doesn't exist.
func InitDB() (*sql.DB, error) {
	var err error
	expandedPath, err := expandPath(dbPath)
	if err != nil {
		return nil, err
	}

	// dbDir := filepath.Dir(expandedPath)
	// // Ensure the directory for the database file exists.
	// if _, err := os.Stat(dbDir); os.IsNotExist(err) {
	// 	if err := os.MkdirAll(dbDir, 0755); err != nil {
	// 		return nil, err
	// 	}
	// }

	DB, err := sql.Open("sqlite3", expandedPath)
	if err != nil {
		return nil, err
	}

	query_vpn_profile := `
	CREATE TABLE IF NOT EXISTS vpn_profile (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		profile TEXT UNIQUE NOT NULL,
		last_connected_at DATETIME
	);`

	_, err = DB.Exec(query_vpn_profile)
	if err != nil {
		DB.Close()
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	query_vpn_user_credential := `
	CREATE TABLE IF NOT EXISTS vpn_user_credential (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		second_password TEXT NOT NULL,
		y_flag TEXT NOT NULL,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = DB.Exec(query_vpn_user_credential)
	if err != nil {
		DB.Close()
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return DB, nil
}

func CleanAllProfile() error {
	// Initialize the database
	db, err := InitDB()
	if err != nil {
		logger.LogError(err, "Error connecting to db")
	}
	defer db.Close()

	// The SQL query to delete all Helm charts.
	query := `DELETE FROM vpn_profile;`

	// Execute the query to delete the Helm charts.
	_, qerr := db.Exec(query)
	if qerr != nil {
		// Return an error if the deletion fails.
		logger.LogError(err, "Error executing query to delete all profiles")
		return err
	}
	// Return nil if the deletion is successful.
	return nil
}

func SetLastConnectedProfile(profile string) error {
	// Initialize the database
	db, err := InitDB()
	if err != nil {
		logger.LogError(err, "Error connecting to db")
	}
	defer db.Close()

	query := `
	INSERT INTO vpn_profile (profile, last_connected_at) VALUES (?, ?)
	ON CONFLICT(profile) DO UPDATE SET last_connected_at=excluded.last_connected_at;
	`
	_, dberr := db.Exec(query, profile, time.Now())
	return dberr
}

func GetLastConnectedProfile() (string, error) {
	// Initialize the database
	db, err := InitDB()
	if err != nil {
		logger.LogError(err, "Error connecting to db")
	}
	defer db.Close()

	query := `SELECT profile FROM vpn_profile ORDER BY last_connected_at DESC LIMIT 1;`
	row := db.QueryRow(query)
	var profile string
	if dberr := row.Scan(&profile); err != nil {
		return "", dberr
	}
	return profile, nil
}
