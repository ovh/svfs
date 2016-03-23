package svfs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/xlucas/swift"
)

const (
	HubicEndpoint = "https://api.hubic.com"
)

var (
	HubicRefreshToken  string
	HubicAuthorization string
)

type hubicCredentials struct {
	Endpoint string `json:"endpoint"`
	Token    string `json:"token"`
}

type hubicToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

// HubicAuth is a swift-compliant authenticatior
// for hubic.
type HubicAuth struct {
	client       http.Client
	credentials  hubicCredentials
	apiToken     hubicToken
	refreshToken string
}

// Request constructs the authentication request.
func (h *HubicAuth) Request(*swift.Connection) (*http.Request, error) {
	form := url.Values{}
	form.Add("refresh_token", HubicRefreshToken)
	form.Add("grant_type", "refresh_token")
	req, err := http.NewRequest("POST", HubicEndpoint+"/oauth/token", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", UserAgent)
	req.Header.Add("Authorization", "Basic "+HubicAuthorization)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Request new API token
	apiResp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer apiResp.Body.Close()

	if apiResp.StatusCode != 200 {
		return nil, fmt.Errorf("Invalid reply from server when fetching hubic API token : %s", apiResp.Status)
	}
	body, err := ioutil.ReadAll(apiResp.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &h.apiToken); err != nil {
		return nil, err
	}

	// Request new keystone token
	req, err = http.NewRequest("GET", HubicEndpoint+"/1.0/account/credentials", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", UserAgent)
	req.Header.Add("Authorization", h.apiToken.TokenType+" "+h.apiToken.AccessToken)
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Invalid reply from server when fetching hubic credentials : %s", resp.Status)
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, &h.credentials); err != nil {
		return nil, err
	}

	return nil, nil
}

// Response reads the authentication response.
func (h *HubicAuth) Response(*http.Response) error {
	return nil
}

// StorageUrl retrieves the swift's storage URL from
// the authentication response.
func (h *HubicAuth) StorageUrl(Internal bool) string {
	return h.credentials.Endpoint
}

// Token retrieves keystone token from the authentication
// response.
func (h *HubicAuth) Token() string {
	return h.credentials.Token
}

// CdnUrl retrives the CDN URL from the authentication
// response.
func (h *HubicAuth) CdnUrl() string {
	return ""
}

var _ swift.Authenticator = (*HubicAuth)(nil)
