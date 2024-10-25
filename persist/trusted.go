package persist

import (
	"encoding/json"
	"fmt"
	"github.com/gambtho/whototrust/model"
	"github.com/gambtho/whototrust/xlog"
	"os"
	"sync"
)

const trustedCharactersFile = "data/trusted_characters.json"

// Mutex for safe concurrent access
var mu sync.Mutex

// LoadTrustedCharacters loads trusted characters and corporations from a file
func LoadTrustedCharacters() (*model.TrustedCharacters, error) {
	mu.Lock()
	defer mu.Unlock()

	file, err := os.Open(trustedCharactersFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &model.TrustedCharacters{
				TrustedCharacters:     []model.TrustedCharacter{},
				TrustedCorporations:   []model.TrustedCorporation{},
				UntrustedCharacters:   []model.TrustedCharacter{},
				UntrustedCorporations: []model.TrustedCorporation{},
			}, nil
		}
		return nil, fmt.Errorf("failed to open trusted characters file: %v", err)
	}
	defer file.Close()

	var trustedData model.TrustedCharacters
	if err := json.NewDecoder(file).Decode(&trustedData); err != nil {
		return nil, fmt.Errorf("failed to decode trusted characters: %v", err)
	}

	return &trustedData, nil
}

// SaveTrustedCharacters saves trusted characters and corporations to a file
func SaveTrustedCharacters(trustedData *model.TrustedCharacters) error {
	mu.Lock()
	defer mu.Unlock()

	file, err := os.Create(trustedCharactersFile)
	if err != nil {
		return fmt.Errorf("failed to create trusted characters file: %v", err)
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(trustedData); err != nil {
		return fmt.Errorf("failed to encode trusted characters: %v", err)
	}

	return nil
}

// AddTrustedCharacter adds a new character to the trusted list
func AddTrustedCharacter(newCharacter model.TrustedCharacter) error {
	trustedData, err := LoadTrustedCharacters()
	if err != nil {
		return fmt.Errorf("failed to load trusted data: %v", err)
	}

	// Check for duplicate
	for _, char := range trustedData.TrustedCharacters {
		if char.CharacterID == newCharacter.CharacterID {
			return nil
		}
	}

	trustedData.TrustedCharacters = append(trustedData.TrustedCharacters, newCharacter)

	return SaveTrustedCharacters(trustedData)
}

// RemoveTrustedCharacter removes a character from the trusted list by CorporationID
func RemoveTrustedCharacter(characterID int64) error {
	trustedData, err := LoadTrustedCharacters()
	if err != nil {
		return fmt.Errorf("failed to load trusted data: %v", err)
	}

	updatedCharacters := make([]model.TrustedCharacter, 0, len(trustedData.TrustedCharacters))
	for _, char := range trustedData.TrustedCharacters {
		if char.CharacterID != characterID {
			updatedCharacters = append(updatedCharacters, char)
		}
	}

	trustedData.TrustedCharacters = updatedCharacters

	return SaveTrustedCharacters(trustedData)
}

// AddTrustedCorporation adds a new corporation to the trusted list
func AddTrustedCorporation(newCorporation model.TrustedCorporation) error {
	trustedData, err := LoadTrustedCharacters()
	if err != nil {
		return fmt.Errorf("failed to load trusted data: %v", err)
	}

	// Check for duplicate
	for _, corp := range trustedData.TrustedCorporations {
		if corp.CorporationID == newCorporation.CorporationID {
			xlog.Logf("corporation already exists - returning success")
			return nil
		}
	}

	trustedData.TrustedCorporations = append(trustedData.TrustedCorporations, newCorporation)

	return SaveTrustedCharacters(trustedData)
}

// RemoveTrustedCorporation removes a corporation from the trusted list by CorporationID
func RemoveTrustedCorporation(id int64) error {
	trustedData, err := LoadTrustedCharacters()
	if err != nil {
		return fmt.Errorf("failed to load trusted data: %v", err)
	}

	updatedCorporations := make([]model.TrustedCorporation, 0, len(trustedData.TrustedCorporations))
	for _, corp := range trustedData.TrustedCorporations {
		if corp.CorporationID != id {
			updatedCorporations = append(updatedCorporations, corp)
		}
	}

	trustedData.TrustedCorporations = updatedCorporations

	return SaveTrustedCharacters(trustedData)
}

func RemoveUntrustedCorporation(id int64) error {
	trustedData, err := LoadTrustedCharacters()
	if err != nil {
		return fmt.Errorf("failed to load trusted data: %v", err)
	}

	updatedCorporations := make([]model.TrustedCorporation, 0, len(trustedData.UntrustedCorporations))
	for _, corp := range trustedData.UntrustedCorporations {
		if corp.CorporationID != id {
			updatedCorporations = append(updatedCorporations, corp)
		}
	}

	trustedData.UntrustedCorporations = updatedCorporations

	return SaveTrustedCharacters(trustedData)
}

// AddUntrustedCorporation adds a new corporation to the untrusted list
func AddUntrustedCorporation(newCorporation model.TrustedCorporation) error {
	trustedData, err := LoadTrustedCharacters()
	if err != nil {
		return fmt.Errorf("failed to load trusted data: %v", err)
	}

	// Check for duplicate
	for _, corp := range trustedData.UntrustedCorporations {
		if corp.CorporationID == newCorporation.CorporationID {
			xlog.Logf("corporation already exists in untrusted list - returning success")
			return nil
		}
	}

	trustedData.UntrustedCorporations = append(trustedData.UntrustedCorporations, newCorporation)

	return SaveTrustedCharacters(trustedData)
}

// AddUntrustedCharacter adds a new character to the untrusted list
func AddUntrustedCharacter(newCharacter model.TrustedCharacter) error {
	trustedData, err := LoadTrustedCharacters()
	if err != nil {
		return fmt.Errorf("failed to load trusted data: %v", err)
	}

	// Check for duplicate
	for _, char := range trustedData.UntrustedCharacters {
		if char.CharacterID == newCharacter.CharacterID {
			xlog.Logf("character already exists in untrusted list - returning success")
			return nil
		}
	}

	trustedData.UntrustedCharacters = append(trustedData.UntrustedCharacters, newCharacter)

	return SaveTrustedCharacters(trustedData)
}

// RemoveUntrustedCharacter removes a character from the untrusted list by CorporationID
func RemoveUntrustedCharacter(characterID int64) error {
	trustedData, err := LoadTrustedCharacters()
	if err != nil {
		return fmt.Errorf("failed to load trusted data: %v", err)
	}

	updatedCharacters := make([]model.TrustedCharacter, 0, len(trustedData.UntrustedCharacters))
	for _, char := range trustedData.UntrustedCharacters {
		if char.CharacterID != characterID {
			updatedCharacters = append(updatedCharacters, char)
		}
	}

	trustedData.UntrustedCharacters = updatedCharacters

	return SaveTrustedCharacters(trustedData)
}
