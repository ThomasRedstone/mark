package confluence

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// RewriteURLForScopedToken rewrites a site-specific Confluence base URL to
// the api.atlassian.com form required for scoped API tokens.
//
// The /wiki path component is located and everything after it is preserved.
// Input:  https://mysite.atlassian.net/wiki (or without /wiki)
// Output: https://api.atlassian.com/ex/confluence/<cloudID>/wiki
func RewriteURLForScopedToken(baseURL, cloudID string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	path := u.Path
	wikiIdx := strings.Index(path, "/wiki")
	var suffix string
	if wikiIdx >= 0 {
		suffix = path[wikiIdx+len("/wiki"):]
	}
	suffix = strings.TrimRight(suffix, "/")

	return "https://api.atlassian.com/ex/confluence/" + cloudID + "/wiki" + suffix, nil
}

// FetchCloudID retrieves the Cloud ID for an Atlassian site by calling the
// anonymous /_edge/tenant_info endpoint. Returns an error if the fetch fails or
// the response does not contain a cloudId.
func FetchCloudID(siteURL string) (string, error) {
	u, err := url.Parse(siteURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	tenantInfoURL := u.Scheme + "://" + u.Host + "/_edge/tenant_info"

	resp, err := http.Get(tenantInfoURL) //nolint:gosec // URL is derived from user-supplied base-url
	if err != nil {
		return "", fmt.Errorf("unable to fetch tenant info from %s: %w", tenantInfoURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("tenant info endpoint %s returned status %d", tenantInfoURL, resp.StatusCode)
	}

	var result struct {
		CloudID string `json:"cloudId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("unable to decode tenant info response: %w", err)
	}

	if result.CloudID == "" {
		return "", fmt.Errorf("tenant info response contained an empty cloudId")
	}

	return result.CloudID, nil
}
