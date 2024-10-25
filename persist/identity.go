package persist

import (
	"fmt"
	"github.com/gambtho/whototrust/xlog"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
)

const dir = "data"

func LoadIdentities(mainIdentity int64) (*Identities, error) {
	if mainIdentity == 0 {
		return nil, fmt.Errorf("logged in user not provided")
	}

	identities := &Identities{Tokens: make(map[int64]oauth2.Token)}

	fileInfo, err := os.Stat(getIdentityFileName(mainIdentity))
	if os.IsNotExist(err) || fileInfo.Size() == 0 {
		xlog.Log("no identity file or file is empty")
		return identities, nil
	}

	err = DecryptData(getIdentityFileName(mainIdentity), identities)
	if err != nil {
		xlog.Log("error in decrypt")
		_ = os.Remove(fileInfo.Name())
		return nil, err
	}

	return identities, nil
}

func GetMainIdentityToken(mainIdentity int64) (oauth2.Token, error) {
	identities, err := LoadIdentities(mainIdentity)
	if err != nil {
		return oauth2.Token{}, fmt.Errorf("unable to retrieve token for main identity")
	}

	return identities.Tokens[mainIdentity], nil
}

// LoadIdentityToken retrieves the token for a given character CorporationID.
func LoadIdentityToken(mainIdentity int64, characterID int64) (oauth2.Token, error) {
	identities, err := LoadIdentities(mainIdentity)
	if err != nil {
		return oauth2.Token{}, fmt.Errorf("unable to retrieve token for character CorporationID %d", characterID)
	}

	token, exists := identities.Tokens[characterID]
	if !exists {
		return oauth2.Token{}, fmt.Errorf("token not found for character CorporationID %d", characterID)
	}

	return token, nil
}

func SaveIdentities(mainIdentity int64, ids *Identities) error {
	if mainIdentity == 0 {
		return fmt.Errorf("no main identity provided")
	}

	return EncryptData(ids, getIdentityFileName(mainIdentity))
}

func getIdentityFileName(mainIdentity int64) string {
	return filepath.Join(dir, fmt.Sprintf("%d_identity.json", mainIdentity))
}

// UpdateIdentities is a helper function that updates identities
func UpdateIdentities(mainIdentity int64, updateFunc func(*Identities) error) error {
	ids, err := LoadIdentities(mainIdentity)

	if err != nil {
		xlog.Log("error in load")

		return err
	}

	if err = updateFunc(ids); err != nil {
		xlog.Log("error in update")
		return err
	}

	if err = SaveIdentities(mainIdentity, ids); err != nil {
		xlog.Log("error in save")
		return err
	}

	return nil
}

func DeleteIdentity(mainIdentity int64) error {
	idFile := getIdentityFileName(mainIdentity)
	return os.Remove(idFile)
}
