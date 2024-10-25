package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gambtho/whototrust/eveapi"
	"github.com/gambtho/whototrust/persist"
	"github.com/gambtho/whototrust/xlog"
)

// Helper function to send JSON responses
func sendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			xlog.Logf("Error encoding JSON response: %v", err)
		}
	}
}

// Helper function to send JSON error responses
func sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": message}); err != nil {
		xlog.Logf("Error encoding JSON error response: %v", err)
	}
}

// AddContactsHandler processes the request to add contacts to the EVE API.
func AddContactsHandler(s *SessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		xlog.Logf("AddContactsHandler invoked")
		var request struct {
			CharacterID int64 `json:"characterID"`
		}

		// Decode the JSON payload
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			xlog.Logf("Error decoding JSON: %v", err)
			sendJSONError(w, "Invalid request payload", http.StatusBadRequest)
			return
		}
		xlog.Logf("Received AddContacts request for CharacterID: %v", request.CharacterID)

		// Retrieve session and token
		session, err := s.Get(r, sessionName)
		if err != nil {
			xlog.Logf("Error retrieving session: %v", err)
			sendJSONError(w, "Session retrieval failed", http.StatusInternalServerError)
			return
		}
		sessionValues := getSessionValues(session)

		token, err := persist.LoadIdentityToken(sessionValues.LoggedInUser, request.CharacterID)
		if err != nil {
			xlog.Logf("Error loading identity token for CharacterID %v: %v", request.CharacterID, err)
			sendJSONError(w, fmt.Sprintf("Character token not found: %v", err), http.StatusInternalServerError)
			return
		}
		xlog.Logf("Loaded token for CharacterID %v: %+v", request.CharacterID, token)

		// Load trusted contacts
		trustedData, err := persist.LoadTrustedCharacters()
		if err != nil {
			xlog.Logf("Error loading trusted contacts: %v", err)
			sendJSONError(w, "Failed to load trusted contacts", http.StatusInternalServerError)
			return
		}
		xlog.Logf("Loaded trusted contacts: %d characters, %d corporations", len(trustedData.TrustedCharacters), len(trustedData.TrustedCorporations))

		// Collect IDs of all trusted contacts
		var contactIDs []int64
		for _, character := range trustedData.TrustedCharacters {
			contactIDs = append(contactIDs, character.CharacterID)
		}
		for _, corporation := range trustedData.TrustedCorporations {
			contactIDs = append(contactIDs, corporation.CorporationID)
		}
		xlog.Logf("Collected %d contact IDs to add for CharacterID %v", len(contactIDs), request.CharacterID)
		if len(contactIDs) == 0 {
			sendJSONResponse(w, http.StatusOK, map[string]string{"message": "No contacts to add"})
			return
		}
		// Use AddContacts to perform the API call
		if err := eveapi.AddContacts(request.CharacterID, &token, contactIDs); err != nil {
			xlog.Logf("Error adding contacts for CharacterID %v: %v", request.CharacterID, err)
			sendJSONError(w, fmt.Sprintf("Error adding contacts: %v", err), http.StatusInternalServerError)
			return
		}

		// Send success response
		xlog.Logf("Contacts added successfully for CharacterID %v", request.CharacterID)
		sendJSONResponse(w, http.StatusOK, map[string]string{"message": "Contacts added successfully"})
	}
}

// DeleteContactsHandler processes the request to delete contacts from the EVE API.
func DeleteContactsHandler(s *SessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			CharacterID int64 `json:"characterID"`
		}

		// Decode the JSON payload
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			xlog.Logf("Error decoding JSON: %v", err)
			sendJSONError(w, "Invalid request payload", http.StatusBadRequest)
			return
		}
		xlog.Logf("Received DeleteContacts request for CharacterID: %v", request.CharacterID)

		// Retrieve session and token
		session, err := s.Get(r, sessionName)
		if err != nil {
			xlog.Logf("Error retrieving session: %v", err)
			sendJSONError(w, "Session retrieval failed", http.StatusInternalServerError)
			return
		}
		sessionValues := getSessionValues(session)

		token, err := persist.LoadIdentityToken(sessionValues.LoggedInUser, request.CharacterID)
		if err != nil {
			xlog.Logf("Error loading identity token for CharacterID %v: %v", request.CharacterID, err)
			sendJSONError(w, fmt.Sprintf("Character token not found: %v", err), http.StatusInternalServerError)
			return
		}
		xlog.Logf("Loaded token for CharacterID %v: %+v", request.CharacterID, token)

		// Load untrusted contacts
		untrustedData, err := persist.LoadTrustedCharacters()
		if err != nil {
			xlog.Logf("Error loading untrusted contacts: %v", err)
			sendJSONError(w, "Failed to load untrusted contacts", http.StatusInternalServerError)
			return
		}
		xlog.Logf("Loaded untrusted contacts: %d characters, %d corporations", len(untrustedData.UntrustedCharacters), len(untrustedData.UntrustedCorporations))

		// Collect IDs of all untrusted contacts
		var contactIDs []int64
		for _, character := range untrustedData.UntrustedCharacters {
			contactIDs = append(contactIDs, character.CharacterID)
		}
		for _, corporation := range untrustedData.UntrustedCorporations {
			contactIDs = append(contactIDs, corporation.CorporationID)
		}
		xlog.Logf("Collected %d contact IDs to delete for CharacterID %v", len(contactIDs), request.CharacterID)

		// Use DeleteContacts to perform the API call
		if err := eveapi.DeleteContacts(request.CharacterID, &token, contactIDs); err != nil {
			xlog.Logf("Error deleting contacts for CharacterID %v: %v", request.CharacterID, err)
			sendJSONError(w, fmt.Sprintf("Error deleting contacts: %v", err), http.StatusInternalServerError)
			return
		}

		// Send success response
		xlog.Logf("Contacts deleted successfully for CharacterID %v", request.CharacterID)
		sendJSONResponse(w, http.StatusOK, map[string]string{"message": "Contacts deleted successfully"})
	}
}
