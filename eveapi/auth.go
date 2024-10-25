package eveapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/gambtho/whototrust/xlog"
)

const (
	tokenURL        = "https://login.eveonline.com/v2/oauth/token"
	requestTimeout  = 10 * time.Second
	contentType     = "application/x-www-form-urlencoded"
	authorization   = "Authorization"
	contentTypeName = "Content-Type"
)

// Custom HTTP client with timeout
var httpClient = &http.Client{
	Timeout: requestTimeout,
}

var oauth2Config *oauth2.Config

// InitializeOAuth initializes the OAuth2 configuration
func InitializeOAuth(clientID, clientSecret, callbackURL string) {
	oauth2Config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  callbackURL,
		Scopes: []string{
			"publicData",
			"esi-search.search_structures.v1",
			"esi-characters.write_contacts.v1",
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://login.eveonline.com/v2/oauth/authorize",
			TokenURL: "https://login.eveonline.com/v2/oauth/token",
		},
	}
}

// GetAuthURL returns the URL for OAuth2 authentication
func GetAuthURL(state string) string {
	return oauth2Config.AuthCodeURL(state)
}

// ExchangeCode exchanges the authorization code for an access token
func ExchangeCode(code string) (*oauth2.Token, error) {
	return oauth2Config.Exchange(context.Background(), code)
}

func RefreshToken(refreshToken string) (*oauth2.Token, error) {
	// Prepare request body data
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	// Create a new request
	req, err := http.NewRequest(http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		xlog.Logf("Failed to create request to refresh token: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set request headers
	req.Header.Add(contentTypeName, contentType)
	req.Header.Add(authorization, "Basic "+base64.StdEncoding.EncodeToString([]byte(oauth2Config.ClientID+":"+oauth2Config.ClientSecret)))

	// Send the request
	resp, err := httpClient.Do(req)
	if err != nil {
		xlog.Logf("Failed to make request to refresh token: %v", err)
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			xlog.Logf("Failed to read response body: %v", readErr)
			return nil, fmt.Errorf("failed to read response body: %w", readErr)
		}
		bodyString := string(bodyBytes)

		xlog.Logf("Received non-OK status code %d for request to refresh token. Response body: %s", resp.StatusCode, bodyString)
		return nil, fmt.Errorf("received non-OK status code %d: %s", resp.StatusCode, bodyString)
	}

	// Decode the response body
	var token oauth2.Token
	if decodeErr := json.NewDecoder(resp.Body).Decode(&token); decodeErr != nil {
		xlog.Logf("Failed to decode response body: %v", decodeErr)
		return nil, fmt.Errorf("failed to decode response: %w", decodeErr)
	}

	return &token, nil
}
