// not used right now
package credentialstore

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
	"github.com/gtank/cryptopasta"
)

const (
	serviceName     = "vpnctl"
	encryptedCreds  = "vpn_credentials.enc"
)

func getFallbackFilePath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".vpnctl", encryptedCreds)
}

// StoreCredential securely stores the credential in keyring or encrypted file
func StoreCredential(username, password string, key []byte) error {
	// Try keyring
	err := keyring.Set(serviceName, username, password)
	if err == nil {
		return nil
	}

	// Fallback: encrypted local file
	var key32 *[32]byte
	if len(key) == 32 {
		key32 = (*[32]byte)(key)
	} else {
		return errors.New("encryption key must be 32 bytes")
	}
	enc, err := cryptopasta.Encrypt([]byte(password), key32)
	if err != nil {
		return err
	}

	// Save to file
	path := getFallbackFilePath()
	_ = os.MkdirAll(filepath.Dir(path), 0700)
	return os.WriteFile(path, enc, 0600)
}

// GetCredential retrieves the credential securely
func GetCredential(username string, key []byte) (string, error) {
	// Try keyring
	pass, err := keyring.Get(serviceName, username)
	if err == nil {
		return pass, nil
	}

	// Fallback: read encrypted file
	data, err := os.ReadFile(getFallbackFilePath())
	if err != nil {
		return "", errors.New("credential not found in keyring or file")
	}

	var key32 *[32]byte
	if len(key) == 32 {
		key32 = (*[32]byte)(key)
	} else {
		return "", errors.New("encryption key must be 32 bytes")
	}
	dec, err := cryptopasta.Decrypt(data, key32)
	if err != nil {
		return "", errors.New("failed to decrypt fallback credentials")
	}

	return string(dec), nil
}
