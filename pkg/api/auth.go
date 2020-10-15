package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	oidc "github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
)

//Auth struct contains oauth2 config and functions
type Auth struct {
	oauth2Config   oauth2.Config
	oauth2Verifier *oidc.IDTokenVerifier
	idToken        *oidc.IDToken
	authType       string
	login          string
}

func NewAuth() *Auth {
	return &Auth{}
}

func (a *Auth) init() error {
	err := a.oauthInit()
	return err
}

func (a *Auth) oauthInit() error {
	ctx := context.Background()

	// Configure an OpenID Connect aware OAuth2 client.
	a.oauth2Config = oauth2.Config{
		ClientID:     os.Getenv("OAUTH2_CLIENT_ID"),
		ClientSecret: os.Getenv("OAUTH2_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("OAUTH2_REDIRECT_URL"),
	}

	if os.Getenv("AUTH_TYPE") == "github" {
		a.oauth2Config.Scopes = []string{"all"}
		a.oauth2Config.Endpoint = oauth2.Endpoint{
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
		}
		a.authType = "github"
	} else {
		provider, err := oidc.NewProvider(ctx, os.Getenv("OAUTH2_URL"))
		if err != nil {
			return err
		}

		// Discovery returns the OAuth2 endpoints.
		a.oauth2Config.Endpoint = provider.Endpoint()

		// verifier
		a.oauth2Verifier = provider.Verifier(&oidc.Config{ClientID: os.Getenv("OAUTH2_CLIENT_ID")})
		// scope
		scopes := os.Getenv("OAUTH2_SCOPES")
		if len(scopes) > 0 {
			a.oauth2Config.Scopes = strings.Split(scopes, " ")
		} else {
			a.oauth2Config.Scopes = []string{oidc.ScopeOpenID, "profile", "email"}
		}

		a.authType = "oidc"
	}

	return nil
}

func (a *Auth) getAuthURL(token string) string {
	return a.oauth2Config.AuthCodeURL(token)
}

func (a *Auth) getToken(code string) (string, error) {
	ctx := context.Background()
	oauth2Token, err := a.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return "", fmt.Errorf("Oauth2 exchange error: %s", err)
	}

	switch a.authType {
	case "oidc":
		// Extract the ID Token from OAuth2 token.
		token, ok := oauth2Token.Extra("id_token").(string)
		if !ok {
			return "", fmt.Errorf("missing token")
		}
		return token, nil
	case "github":
		// return access token
		return oauth2Token.AccessToken, nil
	default:
		return "", fmt.Errorf("Misconfiguration: Auth type not recognized")

	}
}
func (a *Auth) verifyToken(token string) error {
	var err error
	ctx := context.Background()
	switch a.authType {
	case "oidc":
		var claims Claims

		a.idToken, err = a.oauth2Verifier.Verify(ctx, token)
		if err != nil {
			return fmt.Errorf("token verification failed: %s", err)
		}

		if err := a.idToken.Claims(&claims); err != nil {
			return err
		}

		if claims.Email != "" {
			a.login = claims.Email
		} else if claims.Name != "" {
			a.login = claims.Name
		} else {
			return fmt.Errorf("No login found in token claims (email / name is empty)")
		}

		return nil
	case "github":
		var githubUser GitHubUser
		client := &http.Client{}
		req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
		if err != nil {
			return err
		}
		req.Header.Add("Authorization", "token "+token)
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		json.NewDecoder(resp.Body).Decode(&githubUser)

		if githubUser.Message != "" {
			return fmt.Errorf("Github response: " + githubUser.Message)
		}

		a.login = githubUser.Login

		return nil
	default:
		return fmt.Errorf("Misconfiguration: Auth type not recognized")
	}
}

func (a *Auth) getLogin() string {
	return a.login
}
