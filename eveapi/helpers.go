package eveapi

import (
	"errors"
	"fmt"
	"github.com/gambtho/whototrust/xlog"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2"
)

const (
	maxRetries = 5
	baseDelay  = 1 * time.Second
	maxDelay   = 32 * time.Second
)

// retryWithExponentialBackoff retries the given function with exponential backoff
func retryWithExponentialBackoff(operation func() (interface{}, error)) (interface{}, error) {
	var result interface{}
	var err error
	delay := baseDelay

	for i := 0; i < maxRetries; i++ {
		if result, err = operation(); err == nil {
			return result, nil
		}

		var customErr *CustomError
		if !errors.As(err, &customErr) || (customErr.StatusCode != http.StatusServiceUnavailable && customErr.StatusCode != http.StatusGatewayTimeout) && customErr.StatusCode != http.StatusInternalServerError {
			break
		}

		if i == maxRetries-1 {
			break
		}

		jitter := time.Duration(rand.Int63n(int64(delay)))
		time.Sleep(delay + jitter)

		delay *= 2
		if delay > maxDelay {
			delay = maxDelay
		}
	}

	return nil, err
}

// createRequestWithParams builds an HTTP GET request with the specified base URL and query parameters
func createRequestWithParams(baseURL string, params map[string]string, token *oauth2.Token) (*http.Request, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %v", err)
	}

	query := u.Query()
	for key, value := range params {
		query.Set(key, value)
	}
	u.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en")
	req.Header.Set("Cache-Control", "no-cache")

	return req, nil
}

// makeRequest handles basic requests without parameters
func makeRequest(url string, token *oauth2.Token) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		newToken, err := RefreshToken(token.RefreshToken)
		if err != nil {
			return nil, fmt.Errorf("failed to refresh token: %v", err)
		}
		*token = *newToken
		return makeRequest(url, newToken)
	}

	if customErr, exists := httpStatusErrors[resp.StatusCode]; exists {
		return nil, customErr
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, NewCustomError(resp.StatusCode, "failed request")
	}

	return bodyBytes, nil
}

// makeRequestWithParams uses createRequestWithParams to handle requests with parameters
func makeRequestWithParams(baseURL string, params map[string]string, token *oauth2.Token) ([]byte, error) {
	req, err := createRequestWithParams(baseURL, params, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create request with parameters: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		newToken, err := RefreshToken(token.RefreshToken)
		if err != nil {
			return nil, fmt.Errorf("failed to refresh token: %v", err)
		}
		xlog.Logf("token refreshed for %s", baseURL)
		*token = *newToken
		return makeRequestWithParams(baseURL, params, newToken)
	}

	if customErr, exists := httpStatusErrors[resp.StatusCode]; exists {
		xlog.Logf("failed calling %s, %v", baseURL, params)
		return nil, customErr
	}

	if resp.StatusCode != http.StatusOK {
		xlog.Logf("failed calling %s, %v", baseURL, params)
		return nil, NewCustomError(resp.StatusCode, "failed request")
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	return bodyBytes, nil
}

func getResults(address string, token *oauth2.Token, params ...map[string]string) ([]byte, error) {
	var operation func() (interface{}, error)

	if len(params) > 0 && params[0] != nil {
		// If params are provided and not nil, use makeRequestWithParams
		operation = func() (interface{}, error) {
			return makeRequestWithParams(address, params[0], token)
		}
	} else {
		// Otherwise, use makeRequest without parameters
		operation = func() (interface{}, error) {
			return makeRequest(address, token)
		}
	}

	result, err := retryWithExponentialBackoff(operation)
	if err != nil {
		return nil, err
	}

	bodyBytes, ok := result.([]byte)
	if !ok {
		return nil, fmt.Errorf("failed to convert result to byte slice")
	}

	return bodyBytes, nil
}
