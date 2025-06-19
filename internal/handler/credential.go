package handler

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/goo-apps/vpnctl/config"
	"github.com/goo-apps/vpnctl/internal/model"
	"github.com/goo-apps/vpnctl/logger"
	"github.com/zalando/go-keyring"
	"golang.org/x/term"
)

// Save credential securely
func StoreCredential(cred model.CREDENTIAL_FOR_LOGIN) error {
	// encrypt the creds
	encryptedPassword, err := Encrypt(cred.Password, config.KEYRING_ENCRYPTION_KEY)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}
	cred.Password = encryptedPassword
	// Construct single string with newline separators
	encoded := strings.Join([]string{
		cred.Username,
		cred.Password,
		cred.Push,
		cred.YFlag,
	}, "\n")

	// Store in keychain under the profile(vpnctl) name
	if err := keyring.Set(config.KEYRING_SERVICE_NAME, config.KEYRING_SERVICE_NAME, encoded); err != nil {
		return fmt.Errorf("failed to store credentials: %w", err)
	}
	return nil
}

// Get credential securely
func GetCredential() (creds model.CREDENTIAL_FOR_LOGIN, err error) {
	entry, err := keyring.Get(config.KEYRING_SERVICE_NAME, config.KEYRING_SERVICE_NAME)
	if err != nil {
		return model.CREDENTIAL_FOR_LOGIN{}, err
	}

	parts := strings.SplitN(entry, "\n", 4)
	if len(parts) != 4 {
		return model.CREDENTIAL_FOR_LOGIN{}, fmt.Errorf("invalid credential format")
	}

	// decrypt the password 
	decryptedPassword, err := Decrypt(parts[1], config.KEYRING_ENCRYPTION_KEY)
	if err != nil {
		return model.CREDENTIAL_FOR_LOGIN{}, fmt.Errorf("failed to decrypt password: %w", err)
	}
	creds.Username = parts[0]
	creds.Password = decryptedPassword
	creds.Push = parts[2]
	creds.YFlag = parts[3]

	return creds, nil
}

func GetOrPromptCredential() (*model.CREDENTIAL_FOR_LOGIN, error) {
	var credential model.CREDENTIAL_FOR_LOGIN
	
	// Check keyring first
	entry, err := keyring.Get(config.KEYRING_SERVICE_NAME, config.KEYRING_SERVICE_NAME)
	if err == nil {
		parts := strings.SplitN(entry, "\n", 4)
		if len(parts) == 4 {
			// decrypt the password 
			decryptedPassword, err := Decrypt(parts[1], config.KEYRING_ENCRYPTION_KEY)
			if err != nil {
				return &credential, fmt.Errorf("failed to decrypt password: %w", err)
			}
			credential.Username = parts[0]
			credential.Password = decryptedPassword
			credential.Push = "push"
			credential.YFlag = "y"

			return &credential, nil
		}
	}

	// Prompt the user (first time)
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("Enter username for '%s': ", config.KEYRING_SERVICE_NAME)
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Enter password: ")
	bytePassword, _ := term.ReadPassword(int(os.Stdin.Fd()))
	password := string(bytePassword)
	fmt.Println()

	push := "push"
	y_flag := "y"

	encryptedPassword, err := Encrypt(password, config.KEYRING_ENCRYPTION_KEY)
	if err != nil {
		return &model.CREDENTIAL_FOR_LOGIN{}, fmt.Errorf("failed to encrypt password: %w", err)
	}
	password = encryptedPassword

	// Store securely
	cred := username + "\n" + password + "\n" + push + "\n" + y_flag
	if err := keyring.Set(config.KEYRING_SERVICE_NAME, config.KEYRING_SERVICE_NAME, cred); err != nil {
		return &credential, fmt.Errorf("failed to store credentials: %w", err)
	}
	credential.Username = username
	credential.Password = password
	credential.Push = push
	credential.YFlag = y_flag

	return &credential, nil
}

// remove credential from key ring
func RemoveCredential() error {
	err := keyring.Delete(config.KEYRING_SERVICE_NAME, config.KEYRING_SERVICE_NAME)
	if err != nil {
		return err
	}
	logger.Warningf("Your credential has been removed from keyring!!")
	return nil
}

// Encrypt encrypts plain text string with the given key.
func Encrypt(plainText, key string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	cipherText := aesGCM.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// Decrypt decrypts the encrypted string with the given key.
func Decrypt(encryptedText, key string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := aesGCM.NonceSize()
	nonce, cipherText := data[:nonceSize], data[nonceSize:]
	plainText, err := aesGCM.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", err
	}
	return string(plainText), nil
}
