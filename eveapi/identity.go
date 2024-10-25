package eveapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"github.com/gambtho/whototrust/model"
	"github.com/gambtho/whototrust/persist"
	"github.com/gambtho/whototrust/xlog"
)

func PopulateIdentities(userConfig *persist.Identities) (map[int64]model.CharacterData, error) {
	characterData := make(map[int64]model.CharacterData)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for id, token := range userConfig.Tokens {
		wg.Add(1)
		go func(id int64, token oauth2.Token) {
			defer wg.Done()

			charIdentity, err := processIdentity(id, token, userConfig, &mu)
			if err != nil {
				xlog.Logf("Failed to process identity for character %d: %v", id, err)
				return
			}

			mu.Lock()
			characterData[id] = *charIdentity
			mu.Unlock()
		}(id, token)
	}

	wg.Wait()

	return characterData, nil
}

func processIdentity(id int64, token oauth2.Token, userConfig *persist.Identities, mu *sync.Mutex) (*model.CharacterData, error) {
	newToken, err := RefreshToken(token.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token for character %d: %v", id, err)
	}
	token = *newToken

	mu.Lock()
	userConfig.Tokens[id] = token
	mu.Unlock()

	corp, err := GetCharacterCorporation(id, &token)
	if err != nil {
		return nil, fmt.Errorf("failed to get corp for character %d: %v", id, err)
	}

	user, err := GetUserInfo(&token)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %v", err)
	}

	portrait, err := GetCharacterPortrait(id)

	character := model.Character{
		User:          *user,
		CorporationID: int64(corp),
		Portrait:      portrait,
	}

	return &model.CharacterData{
		Token:     token,
		Character: character,
	}, nil
}

func GetUserInfo(token *oauth2.Token) (*model.User, error) {
	if token.AccessToken == "" {
		return nil, fmt.Errorf("no access token provided")
	}

	requestURL := "https://login.eveonline.com/oauth/verify"

	bodyBytes, err := getResults(requestURL, token)
	if err != nil {
		return nil, err
	}

	var user model.User
	if err := json.Unmarshal(bodyBytes, &user); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %v", err)
	}

	return &user, nil
}

func GetCharacterCorporation(characterID int64, token *oauth2.Token) (int32, error) {
	url := fmt.Sprintf("https://esi.evetech.net/latest/characters/%d/?datasource=tranquility", characterID)

	bodyBytes, err := getResults(url, token)
	if err != nil {
		return 0, err
	}

	var character model.CharacterResponse
	if err := json.Unmarshal(bodyBytes, &character); err != nil {
		return 0, fmt.Errorf("failed to decode response body: %v", err)
	}

	return character.CorporationID, nil
}

func GetPublicCharacterData(characterID int64, token *oauth2.Token) (*model.CharacterResponse, error) {
	url := fmt.Sprintf("https://esi.evetech.net/latest/characters/%d/?datasource=tranquility", characterID)

	bodyBytes, err := getResults(url, token)
	if err != nil {
		return nil, err
	}

	var character model.CharacterResponse
	if err := json.Unmarshal(bodyBytes, &character); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %v", err)
	}

	return &character, nil
}

func GetCorpInfo(corporationID int64, token *oauth2.Token) (*model.CorporationInfo, error) {
	url := fmt.Sprintf("https://esi.evetech.net/latest/corporations/%d/", corporationID)
	params := map[string]string{
		"datasource": "tranquility",
	}

	bodyBytes, err := getResults(url, token, params)
	if err != nil {
		return nil, err
	}

	var corp model.CorporationInfo
	if err := json.Unmarshal(bodyBytes, &corp); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %v", err)
	}
	xlog.Logf("%v", corp)

	return &corp, nil
}

func CharacterIDSearch(characterID int64, name string, token *oauth2.Token) (int32, error) {
	baseURL := fmt.Sprintf("https://esi.evetech.net/latest/characters/%d/search/", characterID)
	params := map[string]string{
		"categories": "character",
		"datasource": "tranquility",
		"language":   "en",
		"search":     name,
		"strict":     "true",
	}

	bodyBytes, err := getResults(baseURL, token, params)
	if err != nil {
		return 0, err
	}

	// Parse JSON response
	var result struct {
		Character []int32 `json:"character"`
	}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return 0, fmt.Errorf("failed to parse JSON response: %v", err)
	}

	if len(result.Character) == 0 {
		xlog.Logf("invalid characters returned from esi %v", result)
		return 0, fmt.Errorf("no characterIDs returned for that name")
	}

	tempID := result.Character[0]

	if len(result.Character) > 1 {
		found := false
		for _, charID := range result.Character {
			charData, err := GetPublicCharacterData(int64(charID), token)
			xlog.Logf("%v", charData)
			if err != nil {
				continue
			}
			if charData.Name == name {
				tempID = charID
				found = true
				break
			}
		}
		if !found {
			xlog.Logf("%v returned for  %v", result.Character, name)
			return 0, fmt.Errorf("invalid character IDs returned for that name")
		}
	}

	return tempID, nil
}

func CorporationIDSearch(characterID int64, name string, token *oauth2.Token) (int32, error) {
	baseURL := fmt.Sprintf("https://esi.evetech.net/latest/characters/%d/search/", characterID)

	// Define query parameters including the token
	params := map[string]string{
		"categories": "corporation",
		"datasource": "tranquility",
		"language":   "en",
		"search":     name,
		"strict":     "true",
		"token":      token.AccessToken, // Adding the token as a query parameter
	}

	// Call getResults with the updated params map
	bodyBytes, err := getResults(baseURL, token, params)
	if err != nil {
		return 0, err
	}

	// Parse JSON response
	var result struct {
		Corporation []int32 `json:"corporation"`
	}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return 0, fmt.Errorf("failed to parse JSON response: %v", err)
	}

	if len(result.Corporation) > 1 || len(result.Corporation) == 0 {
		return 0, fmt.Errorf("invalid corporation IDs returned for that name")
	}

	return result.Corporation[0], nil
}

// AddContacts is a helper function to send contacts to the EVE API.
func AddContacts(characterID int64, token *oauth2.Token, contactIDs []int64) error {
	// Prepare JSON payload
	contactIDsJSON, err := json.Marshal(contactIDs)
	if err != nil {
		xlog.Logf("Error encoding contact IDs: %v", err)
		return fmt.Errorf("error encoding contact IDs: %w", err)
	}

	// Build the request URL with query parameters
	baseURL := fmt.Sprintf("https://esi.evetech.net/latest/characters/%d/contacts/", characterID)
	params := url.Values{}
	params.Set("standing", strconv.FormatFloat(5.0, 'f', 1, 64))

	client := &http.Client{}
	req, err := http.NewRequest("POST", baseURL+"?"+params.Encode(), bytes.NewBuffer(contactIDsJSON))
	if err != nil {
		xlog.Logf("Error creating request: %v", err)
		return fmt.Errorf("error creating request: %w", err)
	}

	// Set headers for the request
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Cache-Control", "no-cache")

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		xlog.Logf("Error executing request: %v", err)
		return fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	// Check for a successful response
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		xlog.Logf("Failed to add contacts: status %d", resp.StatusCode)
		return fmt.Errorf("failed to add contacts: received status %d", resp.StatusCode)
	}

	// Decode the response as an array of integers (contact IDs)
	var contacts []int
	if err := json.NewDecoder(resp.Body).Decode(&contacts); err != nil {
		xlog.Logf("Error decoding response body: %v", err)
		return fmt.Errorf("failed to decode response body: %v", err)
	}

	xlog.Logf("Contacts added successfully: %v", contacts)
	return nil
}

// DeleteContacts is a helper function to send contacts to the EVE API.
func DeleteContacts(characterID int64, token *oauth2.Token, contactIDs []int64) error {
	// Prepare JSON payload
	contactIDsJSON, err := json.Marshal(contactIDs)
	if err != nil {
		xlog.Logf("Error encoding contact IDs: %v", err)
		return fmt.Errorf("error encoding contact IDs: %w", err)
	}

	// Build the request URL with query parameters
	baseURL := fmt.Sprintf("https://esi.evetech.net/latest/characters/%d/contacts/", characterID)
	params := url.Values{}
	for _, id := range contactIDs {
		params.Add("contact_ids", strconv.FormatInt(id, 10))
	}
	params.Set("datasource", "tranquility")

	client := &http.Client{}
	req, err := http.NewRequest("DELETE", baseURL+"?"+params.Encode(), bytes.NewBuffer(contactIDsJSON))
	if err != nil {
		xlog.Logf("Error creating request: %v", err)
		return fmt.Errorf("error creating request: %w", err)
	}

	// Set headers for the request
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Cache-Control", "no-cache")

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		xlog.Logf("Error executing request: %v", err)
		return fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	// Check for a successful response
	if resp.StatusCode != http.StatusNoContent {
		xlog.Logf("Failed to delete contacts: status %d", resp.StatusCode)
		return fmt.Errorf("failed to delete contacts: received status %d", resp.StatusCode)
	}

	xlog.Logf("Contacts deleted successfully %v", contactIDs)
	return nil
}

// GetCharacterPortrait retrieves the 64x64 portrait URL for a given characterID.
func GetCharacterPortrait(characterID int64) (string, error) {
	url := fmt.Sprintf("https://esi.evetech.net/latest/characters/%d/portrait/?datasource=tranquility", characterID)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to make request to ESI API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-OK HTTP status: %v", resp.Status)
	}

	var portrait model.CharacterPortrait
	if err := json.NewDecoder(resp.Body).Decode(&portrait); err != nil {
		return "", fmt.Errorf("failed to decode response body: %v", err)
	}

	return portrait.Px64x64, nil
}

func GetAllianceInfo(id int32, token *oauth2.Token) (*model.Alliance, error) {
	url := fmt.Sprintf("https://esi.evetech.net/latest/alliances/%d/", id)
	params := map[string]string{
		"datasource": "tranquility",
	}

	bodyBytes, err := getResults(url, token, params)
	if err != nil {
		return nil, err
	}

	var alliance model.Alliance
	if err := json.Unmarshal(bodyBytes, &alliance); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %v", err)
	}

	xlog.Logf("%v", alliance)
	return &alliance, nil
}
