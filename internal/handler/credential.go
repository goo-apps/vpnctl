package handler

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/goo-apps/vpnctl/internal/model"
	"github.com/zalando/go-keyring"
	"golang.org/x/term"
)

const service = "vpnctl"

// Save credential securely
func StoreCredential(profile, username, password string) error {
	entry := username + "\n" + password
	return keyring.Set(service, profile, entry)
}

// Get credential securely
func GetCredential(profile string) (username, password, push, y_flag string, err error) {
	entry, err := keyring.Get(service, profile)
	if err != nil {
		return "", "", "", "", err
	}

	parts := strings.SplitN(entry, "\n", 4)
	if len(parts) != 4 {
		return "", "", "", "", fmt.Errorf("invalid credential format")
	}
	return parts[0], parts[1], parts[2], parts[3], nil
}

func GetOrPromptCredential(profile string) (model.CREDENTIAL_FOR_LOGIN, error) {
	var credential model.CREDENTIAL_FOR_LOGIN
	// Check keyring first
	entry, err := keyring.Get(service, profile)
	if err == nil {
		parts := strings.SplitN(entry, "\n", 4)
		if len(parts) == 4 {
			credential.Username = parts[0]
			credential.Password = parts[1]
			credential.Push = parts[2]
			credential.YFlag = parts[3]

			return credential, nil
		}
	}

	// Prompt the user (first time)
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Enter username for profile '%s': ", profile)
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Enter password: ")
	bytePassword, _ := term.ReadPassword(int(os.Stdin.Fd()))
	password := string(bytePassword)
	fmt.Println()

	push := "push"
	y_flag := "y"

	// Store securely
	cred := username + "\n" + password + "\n" + push + "\n" + y_flag
	if err := keyring.Set(service, profile, cred); err != nil {
		return credential, fmt.Errorf("failed to store credentials: %w", err)
	}
	credential.Username = username
	credential.Password = password
	credential.Push = push
	credential.YFlag = y_flag

	return credential, nil
}
