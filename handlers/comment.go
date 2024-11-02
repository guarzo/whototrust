package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gambtho/whototrust/persist"
	"github.com/gambtho/whototrust/xlog"
)

// UpdateCommentHandler processes the request to add contacts to the EVE API.
func UpdateCommentHandler(w http.ResponseWriter, r *http.Request) {

	xlog.Logf("Update Comment Handler invoked")
	var request struct {
		ID      int64  `json:"id"`
		Comment string `json:"comment"`
		TableID string `json:"tableId"`
	}

	// Decode the JSON payload
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		xlog.Logf("Error decoding JSON: %v", err)
		sendJSONError(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	xlog.Logf("Received Update Comment request for: %v", request)

	data, err := persist.LoadTrustedCharacters()
	if err != nil {
		xlog.Logf("Error loading trusted characters: %v", err)
		sendJSONError(w, "Error loading trusted characters", http.StatusInternalServerError)
		return
	}

	switch request.TableID {
	case "trusted-characters-table":
		for i, character := range data.TrustedCharacters {
			if character.CharacterID == request.ID {
				data.TrustedCharacters[i].Comment = request.Comment
				break
			}
		}
	case "trusted-corporations-table":
		for i, corporation := range data.TrustedCorporations {
			if corporation.CorporationID == request.ID {
				data.TrustedCorporations[i].Comment = request.Comment
				break
			}
		}
	case "untrusted-characters-table":
		for i, character := range data.UntrustedCharacters {
			if character.CharacterID == request.ID {
				data.UntrustedCharacters[i].Comment = request.Comment
				break
			}
		}
	case "unsanctioned-corporations-table":
		for i, corporation := range data.UntrustedCorporations {
			if corporation.CorporationID == request.ID {
				data.UntrustedCorporations[i].Comment = request.Comment
				break
			}
		}
	default:
		xlog.Logf("table id was not recognized: %v", request.TableID)
		sendJSONError(w, "Error parsing tableID", http.StatusInternalServerError)
		return
	}

	if err := persist.SaveTrustedCharacters(data); err != nil {
		xlog.Logf("Error saving trusted characters: %v", err)
		sendJSONError(w, "Error saving trusted characters", http.StatusInternalServerError)
		return
	}

	sendJSONResponse(w, http.StatusOK, map[string]string{"message": "Comment updated successfully"})
}
