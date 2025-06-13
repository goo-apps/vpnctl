// not used right now
package handler

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/goo-apps/vpnctl/internal/middleware"
	"github.com/goo-apps/vpnctl/internal/model"
	"github.com/goo-apps/vpnctl/logger"
)

var ServerWaitGroup *sync.WaitGroup

// SetCredentialHandler handles the API request to set a credential.
func SetCredentialHandler(w http.ResponseWriter, r *http.Request) {
	var creds model.USER_CREDENTIAL

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Errorf("Failed to read request body", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &creds)
	if err != nil {
		logger.Errorf("Failed to parse JSON", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if creds == (model.USER_CREDENTIAL{}) {
		http.Error(w, "Empty credential payload", http.StatusBadRequest)
		return
	}

	err = middleware.SetCredential(creds)
	if err != nil {
		logger.Errorf("Failed to store credential", err)
		http.Error(w, "Failed to store credential", http.StatusInternalServerError)
		return
	}

	logger.Infof("Credential saved successfully")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Credential saved successfully"}`))

	// Signal main to resume
	if ServerWaitGroup != nil {
		ServerWaitGroup.Done()
	}
}

// GetCredentialHandler handles the API request to get a credential by username.
func GetCredentialHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	creds, err := middleware.GetCredential(username)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Credential not found", http.StatusNotFound)
		} else {
			logger.Errorf("Error getting credential from database", err)
			http.Error(w, "Failed to get credential", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(creds)
	if err != nil {
		logger.Errorf("Error encoding JSON response", err)
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		return
	}
}
