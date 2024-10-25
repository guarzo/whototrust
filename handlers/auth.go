package handlers

import (
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/gambtho/whototrust/eveapi"
	"github.com/gambtho/whototrust/persist"
	"github.com/gambtho/whototrust/xlog"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	state := fmt.Sprintf("main-%d", time.Now().UnixNano())
	url := eveapi.GetAuthURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func AuthCharacterHandler(w http.ResponseWriter, r *http.Request) {
	state := fmt.Sprintf("character-%d", time.Now().UnixNano())
	url := eveapi.GetAuthURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func CallbackHandler(s *SessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")

		token, err := eveapi.ExchangeCode(code)
		if err != nil {
			handleErrorWithRedirect(w, r, fmt.Sprintf("Failed to exchange token for code: %s, state: %s, %v", code, state, err), "/")
			return
		}

		// Get user information
		user, err := eveapi.GetUserInfo(token)
		if err != nil {
			handleErrorWithRedirect(w, r, fmt.Sprintf("Failed to get user info: %v", err), "/")
			return
		}

		session, _ := s.Get(r, sessionName)

		if state[:4] == "main" {
			session.Values[loggedInUser] = user.CharacterID
		}

		if _, ok := session.Values[allAuthenticatedCharacters].([]int64); ok {
			if !slices.Contains(session.Values[allAuthenticatedCharacters].([]int64), user.CharacterID) {
				session.Values[allAuthenticatedCharacters] = append(session.Values[allAuthenticatedCharacters].([]int64), user.CharacterID)
			}
		} else {
			session.Values[allAuthenticatedCharacters] = []int64{user.CharacterID}
		}

		mainIdentity, ok := session.Values[loggedInUser].(int64)
		if !ok || mainIdentity == 0 {
			handleErrorWithRedirect(w, r, fmt.Sprintf("main identity not found, current session: %v", session.Values), "/logout")
			return
		}

		err = persist.UpdateIdentities(mainIdentity, func(userConfig *persist.Identities) error {
			userConfig.Tokens[user.CharacterID] = *token
			return nil
		})

		if err != nil {
			handleErrorWithRedirect(w, r, fmt.Sprintf("Failed to update user config %v", err), "/")
			return
		}

		xlog.Logf("%v logged in", session.Values[allAuthenticatedCharacters])

		session.Save(r, w)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}

func LogoutHandler(s *SessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := s.Get(r, sessionName)
		clearSession(s, w, r)
		session.Save(r, w)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}

func ResetIdentitiesHandler(s *SessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := s.Get(r, sessionName)
		mainIdentity, ok := session.Values[loggedInUser].(int64)

		if !ok || mainIdentity == 0 {
			handleErrorWithRedirect(w, r, "Attempt to reset identities without a main identity", "/logout")
			return
		}

		err := persist.DeleteIdentity(mainIdentity)
		if err != nil {
			xlog.Logf("Failed to delete identity %d: %v", mainIdentity, err)
		}

		http.Redirect(w, r, "/logout", http.StatusSeeOther)
	}
}
