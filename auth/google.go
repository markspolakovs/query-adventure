package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/random"
	"golang.org/x/oauth2"

	"query-adventure/cfg"
)

const googleIssuer = "https://accounts.google.com"
const (
	stateSessionKey = "google-auth-state"
	stateValLength  = 32
)

var authScopes = []string{
	"openid",
	"https://www.googleapis.com/auth/userinfo.profile",
	"https://www.googleapis.com/auth/userinfo.email",
}

type GoogleAuthenticator struct {
	gc       cfg.GoogleCfg
	prov     *oidc.Provider
	verifier *oidc.IDTokenVerifier
	conf     *oauth2.Config
}

func NewGoogleAuthenticator(gc cfg.GoogleCfg) (*GoogleAuthenticator, error) {
	prov, err := oidc.NewProvider(context.TODO(), googleIssuer)
	if err != nil {
		return nil, fmt.Errorf("failed to build provider: %w", err)
	}
	return &GoogleAuthenticator{
		gc:   gc,
		prov: prov,
		verifier: prov.Verifier(&oidc.Config{
			ClientID: gc.ClientID,
		}),
		conf: &oauth2.Config{
			ClientID:     gc.ClientID,
			ClientSecret: gc.ClientSecret,
			RedirectURL:  gc.RedirectURL,
			Scopes: []string{
				"openid",
				"email",
			},
			Endpoint: prov.Endpoint(),
		},
	}, nil
}

func (a *GoogleAuthenticator) MaybeRedirect(ctx echo.Context) (string, error) {
	params := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("scope", strings.Join(authScopes, " ")),
	}
	if hd := a.gc.HostedDomain; hd != "" {
		params = append(params, oauth2.SetAuthURLParam("hd", hd))
	}
	sess, err := session.Get(stateSessionKey, ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get session: %w", err)
	}
	state := random.String(stateValLength, random.Hex)
	sess.Values["state"] = state
	err = sess.Save(ctx.Request(), ctx.Response())
	if err != nil {
		return "", fmt.Errorf("failed to save session: %w", err)
	}
	return a.conf.AuthCodeURL(state, params...), nil
}

func (a *GoogleAuthenticator) Authenticate(ctx echo.Context) (*UserData, error) {
	err := checkSessionState(ctx)
	if err != nil {
		return nil, err
	}

	code := ctx.QueryParam("code")
	if code == "" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "no google code")
	}
	tok, err := a.conf.Exchange(ctx.Request().Context(), code)
	if err != nil {
		var retErr *oauth2.RetrieveError
		if errors.As(err, &retErr) {
			return nil, echo.NewHTTPError(retErr.Response.StatusCode, retErr.Error())
		}
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	rawIDToken, ok := tok.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("no id_token in token")
	}

	idToken, err := a.verifier.Verify(ctx.Request().Context(), rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("invalid ID token: %w", err)
	}

	var claims struct {
		Email      string `json:"email"`
		GivenName  string `json:"given_name"`
		FamilyName string `json:"family_name"`
	}
	err = idToken.Claims(&claims)
	if err != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", err)
	}

	return &UserData{
		FirstName: claims.GivenName,
		LastName:  claims.FamilyName,
		Email:     claims.Email,
	}, nil
}

// checkSessionState verifies that the `state` query parameter matches that stored in the state.
func checkSessionState(ctx echo.Context) error {
	sess, err := session.Get(stateSessionKey, ctx)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	stateParam := ctx.QueryParam("state")
	stateStored := sess.Values["state"]
	if stateStored != stateParam {
		return echo.NewHTTPError(http.StatusBadRequest, "state mismatch")
	}
	delete(sess.Values, "state")
	err = sess.Save(ctx.Request(), ctx.Response())
	if err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}
	return nil
}
