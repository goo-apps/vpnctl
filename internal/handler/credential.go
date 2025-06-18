package handler

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/goo-apps/vpnctl/config"
	"github.com/goo-apps/vpnctl/internal/model"
	"github.com/zalando/go-keyring"
	"golang.org/x/term"
)

// Save credential securely
func StoreCredential(profile string, cred model.CREDENTIAL_FOR_LOGIN) error {
	// Construct single string with newline separators
	encoded := strings.Join([]string{
		cred.Username,
		cred.Password,
		cred.Push,
		cred.YFlag,
	}, "\n")

	// Store in keychain under the profile name
	if err := keyring.Set(config.KEYRING_SERVICE_NAME, profile, encoded); err != nil {
		return fmt.Errorf("failed to store credentials: %w", err)
	}
	return nil
}

// Get credential securely
func GetCredential(profile string) (creds model.CREDENTIAL_FOR_LOGIN, err error) {
	entry, err := keyring.Get(config.KEYRING_SERVICE_NAME, profile)
	if err != nil {
		return model.CREDENTIAL_FOR_LOGIN{}, err
	}

	parts := strings.SplitN(entry, "\n", 4)
	if len(parts) != 4 {
		return model.CREDENTIAL_FOR_LOGIN{}, fmt.Errorf("invalid credential format")
	}
	creds.Username = parts[0]
	creds.Password = parts[1]
	creds.Push = parts[2]
	creds.YFlag = parts[3]

	return creds, nil
}

func GetOrPromptCredential(profile string) (*model.CREDENTIAL_FOR_LOGIN, error) {
	var credential model.CREDENTIAL_FOR_LOGIN
	// Check keyring first
	entry, err := keyring.Get(config.KEYRING_SERVICE_NAME, profile)
	if err == nil {
		parts := strings.SplitN(entry, "\n", 4)
		if len(parts) == 4 {
			credential.Username = parts[0]
			credential.Password = parts[1]
			credential.Push = parts[2]
			credential.YFlag = parts[3]

			return &credential, nil
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
	if err := keyring.Set(config.KEYRING_SERVICE_NAME, profile, cred); err != nil {
		return &credential, fmt.Errorf("failed to store credentials: %w", err)
	}
	credential.Username = username
	credential.Password = password
	credential.Push = push
	credential.YFlag = y_flag

	return &credential, nil
}
