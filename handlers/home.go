package handlers

import (
	"fmt"
	"github.com/gambtho/whototrust/model"
	"github.com/gambtho/whototrust/xlog"
	"net/http"
)

func HomeHandler(s *SessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := s.Get(r, sessionName)
		sessionValues := getSessionValues(session)

		if sessionValues.LoggedInUser == 0 {
			renderLandingPage(w, r)
			return
		}

		storeData, etag, canSkip := checkIfCanSkip(session, sessionValues, r)

		if canSkip {
			renderBaseTemplate(w, r, storeData)
			return
		}

		identities, err := validateIdentities(session, sessionValues, storeData)
		if err != nil {
			handleErrorWithRedirect(w, r, fmt.Sprintf("Failed to validate identities: %v", err), "/logout")
			return
		}

		data := prepareHomeData(sessionValues, identities)

		etag, err = updateStoreAndSession(storeData, data, etag, session, r, w)
		if err != nil {
			xlog.Logf("Failed to update store and session: %v", err)
			return
		}

		renderBaseTemplate(w, r, data)
	}
}

func renderBaseTemplate(w http.ResponseWriter, r *http.Request, data model.HomeData) {
	if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
		handleErrorWithRedirect(w, r, fmt.Sprintf("Failed to render base template: %v", err), "/")
	}
}

func renderLandingPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	data := model.HomeData{Title: Title}
	if err := tmpl.ExecuteTemplate(w, "landing", data); err != nil {
		handleErrorWithRedirect(w, r, fmt.Sprintf("Failed to render landing template: %v", err), "/")
	}
}
