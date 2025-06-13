package middleware

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/goo-apps/vpnctl/logger"

	_ "github.com/mattn/go-sqlite3"

	"github.com/goo-apps/vpnctl/internal/model"
)

const dbPath = "~/.vpnctl/vpnctl.db"

// Function to prompt the user for a username and store it in an environment variable
func GetUsernameFromEnv() string {
	// Check if the username is already set in the DB
	username, err := GetUsernameFromDB()
	if err != nil {
		logger.Errorf("error while fetching user from db: %d", err)
	}
	if username != "" {
		// fmt.Println("Username found in database.")
		return username
	}

	// Prompt the user for a username
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your username to proceed: ")
	username, _ = reader.ReadString('\n')
	username = strings.TrimSpace(username)

	// Validate the username (you can add more complex validation here)
	if username == "" {
		fmt.Println("Username cannot be empty.")
		return GetUsernameFromEnv() // Recursively call the function until a valid username is entered
	}

	// Set the username in the environment variable for future use
	dberr := SetUsernameToDB(username)
	if dberr != nil {
		logger.Errorf("error while setting user to db: %v", err)
	}

	fmt.Println("Username set in environment variable.")
	return username
}

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

	query_vpn_user_env := `
	CREATE TABLE IF NOT EXISTS vpn_user_env (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = DB.Exec(query_vpn_user_env)
	if err != nil {
		DB.Close()
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return DB, nil
}

// func CleanAllProfile() error {
// 	// Initialize the database
// 	db, err := InitDB()
// 	if err != nil {
// 		logger.Errorf(err, "Error connecting to db")
// 	}
// 	defer db.Close()

// 	// The SQL query to delete all Helm charts.
// 	query := `DELETE FROM vpn_profile;`

// 	// Execute the query to delete the Helm charts.
// 	_, qerr := db.Exec(query)
// 	if qerr != nil {
// 		// Return an error if the deletion fails.
// 		logger.Errorf(err, "Error executing query to delete all profiles")
// 		return err
// 	}
// 	// Return nil if the deletion is successful.
// 	return nil
// }

func SetLastConnectedProfile(profile string) error {
	// Initialize the database
	db, err := InitDB()
	if err != nil {
		logger.Errorf("Error connecting to db: %v", err)
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
		logger.Errorf("Error connecting to db: %v", err)
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

// AddCredential adds a new VPN user credential to the vpn_user_credential table.
// It now takes the profile as an argument and uses the Credentials struct.
func SetCredential(creds model.USER_CREDENTIAL) error {
	db, err := InitDB()
	if err != nil {
		logger.Errorf("Error connecting to db: %v", err)
		return err // Important: Return the error
	}
	defer db.Close()

	query := `
    INSERT INTO vpn_user_credential (profile, username, password, second_password, y_flag)
    VALUES (?, ?, ?, ?, ?);
    `

	_, dberr := db.Exec(query, creds.Username, creds.Password, creds.SecondPassword, creds.YFlag)
	if dberr != nil {
		logger.Errorf("Error adding credential: %v", dberr)
		return dberr // Important: Return the error
	}

	return nil
}

// GetCredential retrieves a VPN user credential from the vpn_user_credential table by profile.
// It now returns a Credentials struct.
func GetCredential(username string) (model.USER_CREDENTIAL, error) {
	db, err := InitDB()
	if err != nil {
		logger.Errorf("Error connecting to db: %v", err)
		return model.USER_CREDENTIAL{}, err // Important: Return the error
	}
	defer db.Close()

	query := `
    SELECT username, password, second_password, y_flag
    FROM vpn_user_credential
    WHERE username = ?;
    `

	row := db.QueryRow(query, username)

	var creds model.USER_CREDENTIAL
	creds.Username = username // Set the profile in the struct

	dberr := row.Scan(&creds.Username, &creds.Password, &creds.SecondPassword, &creds.YFlag)
	if dberr != nil {
		if dberr == sql.ErrNoRows {
			return model.USER_CREDENTIAL{}, fmt.Errorf("no credential found for user: %v", username)
		}
		logger.Errorf("Error scanning credential: %v", dberr)
		return model.USER_CREDENTIAL{}, dberr // Important: Return the error
	}

	return creds, nil
}

func SetUsernameToDB(username string) error {
	db, err := InitDB()
	if err != nil {
		logger.Errorf("Error connecting to db: %v", err)
	}
	defer db.Close()

	query := `
	INSERT INTO vpn_user_env (username) VALUES (?)
	ON CONFLICT(username) DO UPDATE SET username=excluded.username;
	`
	_, dberr := db.Exec(query, username, time.Now())
	return dberr
}

func GetUsernameFromDB() (string, error) {
	db, err := InitDB()
	if err != nil {
		logger.Errorf("Error connecting to db: %v", err)
		return "", err // Important: Return the error
	}
	defer db.Close()

	query := `
    SELECT username FROM vpn_user_env;
    `
	row := db.QueryRow(query)
	var username string
	if dberr := row.Scan(&username); dberr != nil {
		return "", dberr
	}

	return username, nil
}
