package vpnctl

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/goo-apps/vpnctl/internal/model"
)

const allReleasesAPI = "https://api.github.com/repos/goo-apps/vpnctl/releases"

func FetchLatestPreOrStableRelease() (*model.GitHubRelease, error) {
	req, err := http.NewRequest("GET", allReleasesAPI, nil)
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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %s", resp.Status)
	}

	var releases []model.GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}

	// Return the first non-draft release (pre-release or stable)
	for _, r := range releases {
		if !r.Draft {
			return &r, nil
		}
	}

	return nil, fmt.Errorf("no valid release found")
}
