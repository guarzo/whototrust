package handlers

import (
	"crypto/rand"
	"net/http"

	"github.com/gorilla/sessions"
)

const (
	lastRefreshTime            = "last_refresh"
	allAuthenticatedCharacters = "authenticated_characters"
	loggedInUser               = "logged_in_user"
	sessionName                = "session"
	previousUserCount          = "previous_user_count"
	previousInputSubbmited     = "previous_input_submitted"
	previousEtagUsed           = "previous_etag_used"
)

type SessionValues struct {
	LastRefreshTime        int64
	LoggedInUser           int64
	PreviousUserCount      int
	PreviousInputSubmitted string
	PreviousEtagUsed       string
}

type SessionService struct {
	store *sessions.CookieStore
}

func getSessionValues(session *sessions.Session) SessionValues {
	s := SessionValues{}

	if val, ok := session.Values[loggedInUser].(int64); ok {
		s.LoggedInUser = val
	}

	if val, ok := session.Values[previousUserCount].(int); ok {
		s.PreviousUserCount = val
	}

	if val, ok := session.Values[previousInputSubbmited].(string); ok {
		s.PreviousInputSubmitted = val
	}

	if val, ok := session.Values[previousEtagUsed].(string); ok {
		s.PreviousEtagUsed = val
	}

	if val, ok := session.Values[lastRefreshTime].(int64); ok {
		s.LastRefreshTime = val
	}

	return s
}

func NewSessionService(secret string) *SessionService {
	return &SessionService{
		store: sessions.NewCookieStore([]byte(secret)),
	}
}

func (s *SessionService) Get(r *http.Request, name string) (*sessions.Session, error) {
	return s.store.Get(r, name)
}

func GenerateSecret() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}
