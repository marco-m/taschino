package release

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/mod/semver"
)

// GitHubLatest queries the GitHub releases API and returns the tag of the
// latest release as an opaque string (that is, it might be or not a valid
// semver string).
func GitHubLatest(owner string, repo string) (string, error) {
	// https://developer.github.com/v3/repos/releases/#get-the-latest-release
	// API: GET /repos/:owner/:repo/releases/latest
	api_url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest",
		owner, repo)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, api_url, nil)
	if err != nil {
		return "", fmt.Errorf("create http request: %w", err)
	}
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("http client Do: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("no release found at %s", api_url)
	}

	type Response struct {
		TagName string `json:"tag_name"`
	}
	var response Response

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&response); err != nil {
		return "", fmt.Errorf("parsing JSON response: %w", err)
	}

	if response.TagName == "" {
		return "", fmt.Errorf("parsing JSON response: missing 'field tag_name'")
	}

	return response.TagName, nil
}

// Compare returns:
// 0 if curV == latestV;
// -1 if curV < latestV;
// +1 if curV > latestV;
// error if curV or latestV are an invalid semver string.
func Compare(curV string, latestV string) (int, error) {
	if !semver.IsValid(curV) {
		return 0,
			fmt.Errorf("installed version is not a valid semver: %s", curV)
	}
	if !semver.IsValid(latestV) {
		return 0,
			fmt.Errorf("latest version is not a valid semver: %s", latestV)
	}
	return semver.Compare(curV, latestV), nil
}
