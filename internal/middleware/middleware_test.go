// Author: rohan.das

// vpnctl - Cross-platform VPN CLI
// Copyright (c) 2025 goo-apps (rohan.das1203@gmail.com)
// Licensed under the MIT License. See LICENSE file for details.
package middleware_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/goo-apps/vpnctl/internal/middleware"
	_ "github.com/mattn/go-sqlite3"

	"github.com/stretchr/testify/assert"
)

// Replace with your actual import path if needed
// import "your_project/internal/middleware"

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	assert.NoError(t, err)

	createTable := `
	CREATE TABLE vpn_profile (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		profile TEXT NOT NULL,
		last_connected_at TEXT NOT NULL
	);
	`
	_, err = db.Exec(createTable)
	assert.NoError(t, err)

	// Insert test data
	_, err = db.Exec(`INSERT INTO vpn_profile (profile, last_connected_at) VALUES (?, ?)`,
		"dev", time.Now().Add(-1*time.Hour).Format(time.RFC3339))
	assert.NoError(t, err)

	_, err = db.Exec(`INSERT INTO vpn_profile (profile, last_connected_at) VALUES (?, ?)`,
		"intra", time.Now().Format(time.RFC3339))
	assert.NoError(t, err)

	return db
}

// TestGetLastConnectedProfile validates that the latest profile is returned
func TestGetLastConnectedProfile(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Overwrite InitDB for test
	originalInitDB := InitDB
	InitDB = func() (*sql.DB, error) {
		return db, nil
	}
	defer func() { InitDB = originalInitDB }()

	profile, err := middleware.GetLastConnectedProfile()
	assert.NoError(t, err)
	assert.Equal(t, "intra", profile)
}

var InitDB = func() (*sql.DB, error) {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".vpnctl", "vpnctl.db")
	return sql.Open("sqlite3", path)
}
