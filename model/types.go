package model

import (
	"time"

	"golang.org/x/oauth2"
)

// CorporationIDs are the IDs of the corporations
var CorporationIDs = []int{98648442, 98670318, 98730557, 98763685, 98743419}

// CharacterIDs are the IDs of the characters
var CharacterIDs = []int{92063989, 96066721, 2114591694, 1959376155, 2115648488, 2121524689, 96180548, 2118868995, 2118016167, 2114311509, 537223062, 2115754172, 629507683, 640170087, 2119887294, 1406208348, 1872552403, 2116275733, 2112148425, 404850015}

type HomeData struct {
	Title                 string
	LoggedIn              bool
	Identities            map[int64]CharacterData
	TabulatorIdentities   []map[string]interface{}
	MainIdentity          int64
	TrustedCharacters     []TrustedCharacter
	TrustedCorporations   []TrustedCorporation
	UntrustedCharacters   []TrustedCharacter
	UntrustedCorporations []TrustedCorporation
}

// Character represents the user information
type Character struct {
	User
	CorporationID int64  `json:"CorporationID"`
	Portrait      string `json:"Portrait"`
}

// CharacterData structure
type CharacterData struct {
	Token oauth2.Token
	Character
}

// User represents the user information returned by the EVE SSO
type User struct {
	CharacterID   int64  `json:"CharacterID"`
	CharacterName string `json:"CharacterName"`
}

type CharacterResponse struct {
	AllianceID     int32     `json:"alliance_id,omitempty"`
	Birthday       time.Time `json:"birthday"`
	BloodlineID    int32     `json:"bloodline_id"`
	CorporationID  int32     `json:"corporation_id"`
	Description    string    `json:"description,omitempty"`
	FactionID      int32     `json:"faction_id,omitempty"`
	Gender         string    `json:"gender"`
	Name           string    `json:"name"`
	RaceID         int32     `json:"race_id"`
	SecurityStatus float64   `json:"security_status,omitempty"`
	Title          string    `json:"title,omitempty"`
}

type TrustedCharacter struct {
	CharacterID     int64     `json:"CharacterID"`
	CharacterName   string    `json:"CharacterName"`
	CorporationID   int64     `json:"CorporationID"`
	CorporationName string    `json:"CorporationName"`
	AddedBy         string    `json:"AddedBy"`
	DateAdded       time.Time `json:"DateAdded"`
	Comment         string    `json:"Comment"`
}

type TrustedCorporation struct {
	CorporationID   int64     `json:"CorporationID"`
	CorporationName string    `json:"CorporationName"`
	AllianceName    string    `json:"AllianceName"`
	AllianceID      int64     `json:"AllianceID"`
	DateAdded       time.Time `json:"DateAdded"`
	AddedBy         string    `json:"AddedBy"`
	Comment         string    `json:"Comment"`
}

type TrustedCharacters struct {
	TrustedCharacters     []TrustedCharacter   `json:"characters"`
	TrustedCorporations   []TrustedCorporation `json:"corporations"`
	UntrustedCharacters   []TrustedCharacter   `json:"untrusted_characters"`
	UntrustedCorporations []TrustedCorporation `json:"untrusted_corporations"`
}

// CharacterSearchResponse represents the array of character IDs returned from the search
type CharacterSearchResponse struct {
	CharacterIDs []int32 `json:"get_characters_character_id_search_character"`
}

type CorporationInfo struct {
	AllianceID    *int32  `json:"alliance_id,omitempty"`     // CorporationID of the alliance, if any
	CEOId         int32   `json:"ceo_id"`                    // CEO CorporationID (required)
	CreatorID     int32   `json:"creator_id"`                // Creator CorporationID (required)
	DateFounded   *string `json:"date_founded,omitempty"`    // Date the corporation was founded
	Description   *string `json:"description,omitempty"`     // CorporationID description
	FactionID     *int32  `json:"faction_id,omitempty"`      // Faction CorporationID, if any
	HomeStationID *int32  `json:"home_station_id,omitempty"` // Home station CorporationID, if any
	MemberCount   int32   `json:"member_count"`              // Number of members (required)
	Name          string  `json:"name"`                      // Full name of the corporation (required)
	Shares        *int64  `json:"shares,omitempty"`          // Number of shares, if any
	TaxRate       float64 `json:"tax_rate"`                  // Tax rate (required, float with max 1.0 and min 0.0)
	Ticker        string  `json:"ticker"`                    // Short name of the corporation (required)
	URL           *string `json:"url,omitempty"`             // CorporationID URL, if any
	WarEligible   *bool   `json:"war_eligible,omitempty"`    // War eligibility, if any
}

// Alliance represents public data about an alliance.
type Alliance struct {
	CreatorCorporationID  int32     `json:"creator_corporation_id"`            // ID of the corporation that created the alliance
	CreatorID             int32     `json:"creator_id"`                        // ID of the character that created the alliance
	DateFounded           time.Time `json:"date_founded"`                      // Date the alliance was founded
	ExecutorCorporationID *int32    `json:"executor_corporation_id,omitempty"` // The executor corporation ID, if available
	FactionID             *int32    `json:"faction_id,omitempty"`              // Faction ID this alliance is fighting for, if enlisted in factional warfare
	Name                  string    `json:"name"`                              // The full name of the alliance
	Ticker                string    `json:"ticker"`                            // The short name (ticker) of the alliance
}

type CharacterPortrait struct {
	Px128x128 string `json:"px128x128"`
	Px256x256 string `json:"px256x256"`
	Px512x512 string `json:"px512x512"`
	Px64x64   string `json:"px64x64"`
}
