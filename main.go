package main

import (
	"encoding/base64"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/gambtho/whototrust/eveapi"
	"github.com/gambtho/whototrust/handlers"
	"github.com/gambtho/whototrust/persist"
	"github.com/gambtho/whototrust/xlog"
)

var version = "0.0.0"

func main() {
	xlog.Logf("Starting application, version %s", version)

	// Read environment variables
	clientID := os.Getenv("EVE_CLIENT_ID")
	clientSecret := os.Getenv("EVE_CLIENT_SECRET")
	callbackURL := os.Getenv("EVE_CALLBACK_URL")

	if clientID == "" || clientSecret == "" || callbackURL == "" {
		log.Fatalf("EVE_CLIENT_ID, EVE_CLIENT_SECRET, and EVE_CALLBACK_URL must be set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	var key []byte
	var err error
	secret := os.Getenv("SECRET_KEY")
	if secret == "" {
		key, err = handlers.GenerateSecret()
		if err != nil {
			log.Fatalf("Failed to generate key: %v", err)
		}
		secret = base64.StdEncoding.EncodeToString(key)

		log.Printf("Generated key: %s -- this should only be used for testing", secret)
	} else {
		key, err = base64.StdEncoding.DecodeString(secret)
		if err != nil {
			log.Fatalf("Failed to decode key: %v", err)
		}
	}

	// Initialize configuration directory
	if err = persist.Initialize(key); err != nil {
		log.Fatalf("Failed to initialize identity: %v", err)
	}

	// Initialize OAuth2 configuration
	eveapi.InitializeOAuth(clientID, clientSecret, callbackURL)

	sessionStore := handlers.NewSessionService(secret)

	// Router setup
	r := mux.NewRouter()

	// utility functions
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	r.HandleFunc("/callback/", handlers.CallbackHandler(sessionStore))

	// user functions
	r.HandleFunc("/", handlers.HomeHandler(sessionStore))
	r.HandleFunc("/login", handlers.LoginHandler)
	r.HandleFunc("/auth-character", handlers.AuthCharacterHandler)
	r.HandleFunc("/logout", handlers.LogoutHandler(sessionStore))

	r.HandleFunc("/update-comment", handlers.UpdateCommentHandler)

	r.HandleFunc("/validate-and-add-trusted-character", handlers.AddTrustedCharacterHandler(sessionStore)) // POST
	r.HandleFunc("/remove-trusted-character", handlers.RemoveTrustedCharacterHandler)

	r.HandleFunc("/validate-and-add-trusted-corporation", handlers.AddTrustedCorporationHandler(sessionStore)) // POST
	r.HandleFunc("/remove-trusted-corporation", handlers.RemoveTrustedCorporationHandler)

	r.HandleFunc("/add-contacts", handlers.AddContactsHandler(sessionStore))
	r.HandleFunc("/delete-contacts", handlers.DeleteContactsHandler(sessionStore))

	r.HandleFunc("/validate-and-add-untrusted-character", handlers.AddUntrustedCharacterHandler(sessionStore)) // POST
	r.HandleFunc("/remove-untrusted-character", handlers.RemoveUntrustedCharacterHandler)

	r.HandleFunc("/validate-and-add-untrusted-corporation", handlers.AddUntrustedCorporationHandler(sessionStore)) // POST
	r.HandleFunc("/remove-untrusted-corporation", handlers.RemoveUntrustedCorporationHandler)

	// admin routes
	r.HandleFunc("/reset-identities", handlers.ResetIdentitiesHandler(sessionStore))

	http.Handle("/", r)

	xlog.Logf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
