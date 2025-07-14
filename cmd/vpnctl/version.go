package vpnctl

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/goo-apps/vpnctl/internal/model"
)

const repoAPI = "https://api.github.com/repos/goo-apps/vpnctl/releases/latest"

func FetchLatestRelease() (*model.GitHubRelease, error) {
	req, err := http.NewRequest("GET", repoAPI, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "vpnctl-release-checker")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned status %s", resp.Status)
	}

	var release model.GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}
