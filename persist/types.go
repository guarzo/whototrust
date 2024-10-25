package persist

import (
	"golang.org/x/oauth2"
)

type Identities struct {
	MainIdentity string                 `json:"main_identity"`
	Tokens       map[int64]oauth2.Token `json:"identities"`
}
