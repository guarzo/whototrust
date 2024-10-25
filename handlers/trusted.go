// handlers/trusted_characters.go
package handlers

import (
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gambtho/whototrust/eveapi"
	"github.com/gambtho/whototrust/model"
	"github.com/gambtho/whototrust/persist"
	"github.com/gambtho/whototrust/xlog"
)

// ErrorResponse represents a JSON-formatted error message.
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse represents a JSON-formatted success message.
type SuccessResponse struct {
	Message string `json:"message"`
}

// EntityData holds the resolved data for an entity, including corporation details if applicable.
type EntityData struct {
	ID              int64
	Name            string
	CorporationID   int64
	CorporationName string
	AllianceID      int64
	AllianceName    string
}

// Helper function to send JSON responses.
func writeJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// If encoding fails, log the error and send a generic error message.
		xlog.Logf("Failed to encode JSON response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Internal Server Error"})
	}
}

// Helper function to send JSON-formatted error messages with identifier context.
func writeJSONError(w http.ResponseWriter, message string, identifier string, statusCode int) {
	fullMessage := message
	if identifier != "" {
		fullMessage = fmt.Sprintf("%s: %s", message, identifier)
	}
	writeJSONResponse(w, ErrorResponse{Error: fullMessage}, statusCode)
}

// Helper function to retrieve main identity and token from session.
func getSessionIdentity(s *SessionService, r *http.Request) (int64, oauth2.Token, error) {
	session, err := s.Get(r, sessionName)
	if err != nil {
		xlog.Logf("Session retrieval error: %v", err)
		return 0, oauth2.Token{}, fmt.Errorf("failed to retrieve session")
	}

	mainIdentity, ok := session.Values[loggedInUser].(int64)
	if !ok || mainIdentity == 0 {
		errorMessage := fmt.Sprintf("Main identity not found, current session: %v", session.Values)
		xlog.Logf(errorMessage)
		return 0, oauth2.Token{}, fmt.Errorf("main identity not found")
	}

	// Retrieve token for main identity.
	token, err := persist.GetMainIdentityToken(mainIdentity)
	if err != nil {
		errorMessage := fmt.Sprintf("Error retrieving token for main identity: %v", err)
		xlog.Logf(errorMessage)
		return 0, oauth2.Token{}, fmt.Errorf("failed to retrieve token")
	}

	return mainIdentity, token, nil
}

// Helper function to parse and resolve the identifier.
func resolveIdentifier(identifier string, entityType string, mainIdentity int64, token *oauth2.Token) (EntityData, error) {
	// Trim spaces.
	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return EntityData{}, fmt.Errorf("identifier is empty")
	}

	// Check if identifier is numeric.
	if id, err := strconv.ParseInt(identifier, 10, 64); err == nil {
		if id <= 0 {
			return EntityData{}, fmt.Errorf("identifier must be a positive number")
		}
		return EntityData{ID: id, Name: ""}, nil
	}

	// Else, treat as name and resolve to ID.
	var resolvedID int32
	var err error
	if entityType == "character" {
		xlog.Logf("Resolving character name to ID: %v", identifier)
		resolvedID, err = eveapi.CharacterIDSearch(mainIdentity, identifier, token)
	} else if entityType == "corporation" {
		xlog.Logf("Resolving corporation name to ID: %v", identifier)
		resolvedID, err = eveapi.CorporationIDSearch(mainIdentity, identifier, token)
	} else {
		return EntityData{}, fmt.Errorf("unknown entity type: %s", entityType)
	}

	if err != nil {
		return EntityData{}, fmt.Errorf("failed to resolve name to ID: %v, try adding by ID instead", err)
	}

	if resolvedID <= 0 {
		return EntityData{}, fmt.Errorf("resolved ID is invalid for identifier: %s", identifier)
	}

	return EntityData{ID: int64(resolvedID), Name: identifier}, nil
}

// Helper function to fetch entity data based on type and ID.
func fetchEntityData(entityType string, data EntityData, token *oauth2.Token) (EntityData, error) {
	if entityType == "character" {
		xlog.Logf("Fetching character data for ID: %v", data.ID)
		characterData, err := eveapi.GetPublicCharacterData(data.ID, token)
		if err != nil {
			return EntityData{}, fmt.Errorf("error retrieving character data: %v", err)
		}

		// Fetch corporation name.
		xlog.Logf("Fetching corporation data for CharacterID: %v", data.ID)
		corpID, err := eveapi.GetCharacterCorporation(data.ID, token)
		if err != nil {
			return EntityData{}, fmt.Errorf("error retrieving character's corporation ID: %v", err)
		}

		corp, err := eveapi.GetCorpInfo(int64(corpID), token)
		if err != nil {
			return EntityData{}, fmt.Errorf("error retrieving corporation info: %v", err)
		}

		// Assign fetched data to EntityData
		data.Name = characterData.Name
		data.CorporationID = int64(corpID)
		data.CorporationName = corp.Name

		return data, nil

	} else if entityType == "corporation" {
		xlog.Logf("Fetching corporation name for ID: %v", data.ID)
		corp, err := eveapi.GetCorpInfo(data.ID, token)
		if err != nil {
			return EntityData{}, fmt.Errorf("error retrieving corporation name: %v", err)
		}

		allianceID := corp.AllianceID
		if allianceID != nil {
			alliance, err := eveapi.GetAllianceInfo(*allianceID, token)
			if err != nil {
				return EntityData{}, fmt.Errorf("error retrieving alliance info: %v", err)
			}
			data.AllianceName = alliance.Name
			data.AllianceID = int64(*allianceID)
		}

		data.Name = corp.Name
		return data, nil
	}

	return EntityData{}, fmt.Errorf("unknown entity type: %s", entityType)
}

func handleAddEntity(s *SessionService, w http.ResponseWriter, r *http.Request, trustStatus string, entityType string) {
	// Decode request body to accept 'identifier'.
	var request struct {
		Identifier string `json:"identifier"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		xlog.Logf("Bad request payload: %v", err)
		writeJSONError(w, "Invalid request payload", request.Identifier, http.StatusBadRequest)
		return
	}

	xlog.Logf("Adding %s %s with identifier: %v", trustStatus, entityType, request.Identifier)

	// Retrieve main identity and token regardless of trustStatus.
	mainIdentity, token, err := getSessionIdentity(s, r)
	if err != nil {
		writeJSONError(w, "Authentication required", request.Identifier, http.StatusUnauthorized)
		return
	}

	// Resolve identifier.
	resolvedData, err := resolveIdentifier(request.Identifier, entityType, mainIdentity, &token)
	if err != nil {
		xlog.Logf("Identifier resolution error: %v", err)
		writeJSONError(w, "Identifier resolution failed", request.Identifier, http.StatusBadRequest)
		return
	}

	// Fetch entity data.
	fetchedData, err := fetchEntityData(entityType, resolvedData, &token)
	if err != nil {
		xlog.Logf("Entity data fetching error: %v", err)
		writeJSONError(w, "Entity data retrieval failed", request.Identifier, http.StatusInternalServerError)
		return
	}

	// Get the character name of the main identity for 'AddedBy' field.
	var addedByName string
	if trustStatus == "trusted" || trustStatus == "untrusted" {
		addedByCharacter, err := eveapi.GetPublicCharacterData(mainIdentity, &token)
		if err != nil {
			xlog.Logf("Error retrieving character data for AddedBy field: %v", err)
			writeJSONError(w, "AddedBy character validation failed", request.Identifier, http.StatusInternalServerError)
			return
		}
		addedByName = addedByCharacter.Name
	}

	// Create the corresponding model based on trustStatus and entityType.
	switch {
	case trustStatus == "trusted" && entityType == "character":
		trustedCharacter := model.TrustedCharacter{
			CharacterID:     fetchedData.ID,
			CharacterName:   fetchedData.Name,
			CorporationID:   fetchedData.CorporationID,
			CorporationName: fetchedData.CorporationName,
			AddedBy:         addedByName,
			DateAdded:       time.Now(),
		}

		xlog.Logf("Adding new trusted character: %+v", trustedCharacter)

		// Persist the trusted character.
		if err := persist.AddTrustedCharacter(trustedCharacter); err != nil {
			xlog.Logf("Error saving trusted character: %v", err)
			writeJSONError(w, "Failed to save trusted character", request.Identifier, http.StatusInternalServerError)
			return
		}

		// Respond with the new trusted character data.
		writeJSONResponse(w, trustedCharacter, http.StatusOK)

	case trustStatus == "trusted" && entityType == "corporation":
		trustedCorporation := model.TrustedCorporation{
			CorporationID:   fetchedData.ID,
			CorporationName: fetchedData.Name,
			AllianceName:    fetchedData.AllianceName,
			AllianceID:      fetchedData.AllianceID,
			DateAdded:       time.Now(),
			AddedBy:         addedByName,
		}

		xlog.Logf("Adding new trusted corporation: %+v", trustedCorporation)

		// Persist the trusted corporation.
		if err := persist.AddTrustedCorporation(trustedCorporation); err != nil {
			xlog.Logf("Error saving trusted corporation: %v", err)
			writeJSONError(w, "Failed to save trusted corporation", request.Identifier, http.StatusInternalServerError)
			return
		}

		// Respond with the new trusted corporation data.
		writeJSONResponse(w, trustedCorporation, http.StatusOK)

	case trustStatus == "untrusted" && entityType == "character":
		untrustedCharacter := model.TrustedCharacter{ // Correct model
			CharacterID:     fetchedData.ID,
			CharacterName:   fetchedData.Name,
			CorporationName: fetchedData.CorporationName,
			CorporationID:   fetchedData.CorporationID,
			DateAdded:       time.Now(),
			AddedBy:         addedByName,
		}

		xlog.Logf("Adding new untrusted character: %+v", untrustedCharacter)

		// Persist the untrusted character.
		if err := persist.AddUntrustedCharacter(untrustedCharacter); err != nil {
			xlog.Logf("Error saving untrusted character: %v", err)
			writeJSONError(w, "Failed to save untrusted character", request.Identifier, http.StatusInternalServerError)
			return
		}

		// Respond with the new untrusted character data.
		writeJSONResponse(w, untrustedCharacter, http.StatusOK)

	case trustStatus == "untrusted" && entityType == "corporation":
		untrustedCorporation := model.TrustedCorporation{ // Correct model
			CorporationID:   fetchedData.ID,
			CorporationName: fetchedData.Name,
			AllianceName:    fetchedData.AllianceName,
			AllianceID:      fetchedData.AllianceID,
			DateAdded:       time.Now(),
			AddedBy:         addedByName,
		}

		xlog.Logf("Adding new untrusted corporation: %+v", untrustedCorporation)

		// Persist the untrusted corporation.
		if err := persist.AddUntrustedCorporation(untrustedCorporation); err != nil {
			xlog.Logf("Error saving untrusted corporation: %v", err)
			writeJSONError(w, "Failed to save untrusted corporation", request.Identifier, http.StatusInternalServerError)
			return
		}

		// Respond with the new untrusted corporation data.
		writeJSONResponse(w, untrustedCorporation, http.StatusOK)

	default:
		xlog.Logf("Unsupported trustStatus or entityType: %s, %s", trustStatus, entityType)
		writeJSONError(w, "Unsupported operation", request.Identifier, http.StatusBadRequest)
	}
}

// Generic function to handle removing entities.
func handleRemoveEntity(w http.ResponseWriter, r *http.Request, trustStatus string, entityType string) {
	// Decode request body to accept 'identifier'.
	var request struct {
		Identifier string `json:"identifier"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		xlog.Logf("Bad request payload: %v", err)
		writeJSONError(w, "Invalid request payload", request.Identifier, http.StatusBadRequest)
		return
	}

	xlog.Logf("Removing %s %s with identifier: %v", trustStatus, entityType, request.Identifier)

	// Parse identifier.
	resolvedData, err := resolveIdentifier(request.Identifier, entityType, 0, nil)
	if err != nil {
		xlog.Logf("Identifier resolution error: %v", err)
		writeJSONError(w, "Identifier resolution failed", request.Identifier, http.StatusBadRequest)
		return
	}

	// For removal, we expect only IDs, not names.
	if resolvedData.ID == 0 {
		writeJSONError(w, "Identifier must be a valid ID for removal", request.Identifier, http.StatusBadRequest)
		return
	}

	// Perform removal based on trustStatus and entityType.
	switch {
	case trustStatus == "trusted" && entityType == "character":
		err = persist.RemoveTrustedCharacter(resolvedData.ID)
		if err != nil {
			xlog.Logf("Error removing trusted character: %v", err)
			writeJSONError(w, "Failed to remove trusted character", request.Identifier, http.StatusInternalServerError)
			return
		}
		writeJSONResponse(w, SuccessResponse{Message: "Trusted character removed successfully"}, http.StatusOK)

	case trustStatus == "trusted" && entityType == "corporation":
		err = persist.RemoveTrustedCorporation(resolvedData.ID)
		if err != nil {
			xlog.Logf("Error removing trusted corporation: %v", err)
			writeJSONError(w, "Failed to remove trusted corporation", request.Identifier, http.StatusInternalServerError)
			return
		}
		writeJSONResponse(w, SuccessResponse{Message: "Trusted corporation removed successfully"}, http.StatusOK)

	case trustStatus == "untrusted" && entityType == "character":
		err = persist.RemoveUntrustedCharacter(resolvedData.ID)
		if err != nil {
			xlog.Logf("Error removing untrusted character: %v", err)
			writeJSONError(w, "Failed to remove untrusted character", request.Identifier, http.StatusInternalServerError)
			return
		}
		writeJSONResponse(w, SuccessResponse{Message: "Untrusted character removed successfully"}, http.StatusOK)

	case trustStatus == "untrusted" && entityType == "corporation":
		err = persist.RemoveUntrustedCorporation(resolvedData.ID)
		if err != nil {
			xlog.Logf("Error removing untrusted corporation: %v", err)
			writeJSONError(w, "Failed to remove untrusted corporation", request.Identifier, http.StatusInternalServerError)
			return
		}
		writeJSONResponse(w, SuccessResponse{Message: "Untrusted corporation removed successfully"}, http.StatusOK)

	default:
		xlog.Logf("Unsupported trustStatus or entityType: %s, %s", trustStatus, entityType)
		writeJSONError(w, "Unsupported operation", request.Identifier, http.StatusBadRequest)
	}
}

// AddTrustedCharacterHandler validates and adds a trusted character.
func AddTrustedCharacterHandler(s *SessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleAddEntity(s, w, r, "trusted", "character")
	}
}

// RemoveTrustedCharacterHandler removes a trusted character by identifier.
func RemoveTrustedCharacterHandler(w http.ResponseWriter, r *http.Request) {
	handleRemoveEntity(w, r, "trusted", "character")
}

// AddTrustedCorporationHandler validates and adds a trusted corporation.
func AddTrustedCorporationHandler(s *SessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleAddEntity(s, w, r, "trusted", "corporation")
	}
}

// RemoveTrustedCorporationHandler removes a trusted corporation by identifier.
func RemoveTrustedCorporationHandler(w http.ResponseWriter, r *http.Request) {
	handleRemoveEntity(w, r, "trusted", "corporation")
}

// AddUntrustedCharacterHandler validates and adds an untrusted character.
func AddUntrustedCharacterHandler(s *SessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleAddEntity(s, w, r, "untrusted", "character")
	}
}

// RemoveUntrustedCharacterHandler removes an untrusted character by identifier.
func RemoveUntrustedCharacterHandler(w http.ResponseWriter, r *http.Request) {
	handleRemoveEntity(w, r, "untrusted", "character")
}

// AddUntrustedCorporationHandler validates and adds an untrusted corporation.
func AddUntrustedCorporationHandler(s *SessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleAddEntity(s, w, r, "untrusted", "corporation")
	}
}

// RemoveUntrustedCorporationHandler removes an untrusted corporation by identifier.
func RemoveUntrustedCorporationHandler(w http.ResponseWriter, r *http.Request) {
	handleRemoveEntity(w, r, "untrusted", "corporation")
}
