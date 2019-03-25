package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	oidc "github.com/coreos/go-oidc"
	"github.com/gorilla/csrf"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

/*
 * Config has a Port property
 */
type Config struct {
	Port string
}

type server struct {
	config         Config
	oauth2Config   oauth2.Config
	oauth2Verifier *oidc.IDTokenVerifier
	sessionStore   *sessions.CookieStore
}

type response struct {
	Message string `json:"message"`
}
type errorResponse struct {
	Message string `json:"message"`
}

/*
 * NewServer initializes new server
 */
func NewServer(conf Config) *server {
	return &server{
		config: conf,
	}
}
func (s *server) Start() {
	r := mux.NewRouter()
	r.HandleFunc("/", s.homeHandler)
	r.HandleFunc("/login", s.loginHandler)
	r.HandleFunc("/callback", s.callbackHandler)
	r.HandleFunc("/ovpnconfig", s.ovpnConfigHandler)
	http.Handle("/", r)

	// initialize oidc
	err := s.oauthInit()
	if err != nil {
		log.Fatalf("Could not initialize oauth2: %s", err)
	}

	// initialize session store
	s.sessionStore = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

	// enable csrf
	CSRF := csrf.Protect([]byte(os.Getenv("CSRF_KEY")))

	// enable logging
	loggedRouter := handlers.LoggingHandler(os.Stdout, CSRF(r))

	// start server
	fmt.Printf("Starting server on port %s\n", s.config.Port)
	log.Fatal(http.ListenAndServe(":"+s.config.Port, loggedRouter))
}

func (s *server) oauthInit() error {
	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, os.Getenv("OAUTH2_URL"))
	if err != nil {
		return err
	}

	// Configure an OpenID Connect aware OAuth2 client.
	s.oauth2Config = oauth2.Config{
		ClientID:     os.Getenv("OAUTH2_CLIENT_ID"),
		ClientSecret: os.Getenv("OAUTH2_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("OAUTH2_REDIRECT_URL"),

		// Discovery returns the OAuth2 endpoints.
		Endpoint: provider.Endpoint(),

		// "openid" is a required scope for OpenID Connect flows.
		Scopes: []string{oidc.ScopeOpenID, "profile", "email"},
	}

	s.oauth2Verifier = provider.Verifier(&oidc.Config{ClientID: os.Getenv("OAUTH2_CLIENT_ID")})

	return nil
}

func (s *server) homeHandler(w http.ResponseWriter, r *http.Request) {
	var response response
	response.Message = "running"
	json.NewEncoder(w).Encode(response)

}

func (s *server) loginHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, s.oauth2Config.AuthCodeURL(csrf.Token(r)), http.StatusFound)
}
func (s *server) callbackHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := s.sessionStore.Get(r, "token-session")

	ctx := context.Background()
	oauth2Token, err := s.oauth2Config.Exchange(ctx, r.URL.Query().Get("code"))
	if err != nil {
		json.NewEncoder(w).Encode(errorResponse{Message: "Oauth2 exchange error"})
		return
	}

	// Extract the ID Token from OAuth2 token.
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		json.NewEncoder(w).Encode(errorResponse{Message: "missing token"})
		return
	}

	// save token
	session.Values["token"] = rawIDToken
	session.Save(r, w)

	// Parse and verify ID Token payload.
	_, err = s.oauth2Verifier.Verify(ctx, rawIDToken)
	if err != nil {
		// handle error
		json.NewEncoder(w).Encode(errorResponse{Message: "token verification failed"})
		return
	}
}

func (s *server) ovpnConfigHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := s.sessionStore.Get(r, "token-session")
	ctx := context.Background()
	idToken, err := s.oauth2Verifier.Verify(ctx, session.Values["token"].(string))
	if err != nil {
		// handle error
		json.NewEncoder(w).Encode(errorResponse{Message: "Unauthorized - token verification failed"})
		return
	}

	// Extract custom claims
	var claims struct {
		Email    string `json:"email"`
		Verified bool   `json:"email_verified"`
	}
	if err := idToken.Claims(&claims); err != nil {
		json.NewEncoder(w).Encode(errorResponse{Message: "claims error"})
		return
	}

	// retrieve CA key / crt
	s3 := NewS3()
	caKey, err := s3.getObject(os.Getenv("S3_BUCKET"), os.Getenv("S3_PREFIX")+"/ca.crt")
	if err != nil {
		json.NewEncoder(w).Encode(errorResponse{Message: "ca.crt download error: " + err.Error()})
	}
	caCert, err := s3.getObject(os.Getenv("S3_BUCKET"), os.Getenv("S3_PREFIX")+"/ca.key")
	if err != nil {
		json.NewEncoder(w).Encode(errorResponse{Message: "ca.crt download error: " + err.Error()})
	}
	// create new cert
	c := NewCert()
	parsedCaCert, err := c.readCert(caCert.String())
	if err != nil {
		json.NewEncoder(w).Encode(errorResponse{Message: "Parsed CA cert Error: " + err.Error()})
		return
	}
	parsedCaKey, err := c.readPrivateKey(caKey.String())
	if err != nil {
		json.NewEncoder(w).Encode(errorResponse{Message: "Parsed CA key Error: " + err.Error()})
		return
	}
	clientCert, clientKey, err := c.createClientCert(parsedCaCert, parsedCaKey, "test-subject")
	if err != nil {
		json.NewEncoder(w).Encode(errorResponse{Message: "Create Cert error: " + err.Error()})
		return
	}

	// output openvpn config
	ovpnConfig, err := s3.getObject(os.Getenv("S3_BUCKET"), os.Getenv("S3_PREFIX")+"/openvpn-client.conf")
	strOvpnConfig := ovpnConfig.String()
	if err != nil {
		json.NewEncoder(w).Encode(errorResponse{Message: "openvpn-client.conf download error: " + err.Error()})
	}
	strOvpnConfig = strings.Replace(strOvpnConfig, "[CERT]", clientCert.String(), -1)
	strOvpnConfig = strings.Replace(strOvpnConfig, "[KEY]", clientKey.String(), -1)
	fmt.Fprintf(w, strOvpnConfig)
}
