package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type Authenticator interface {
	MaybeRedirect(ctx echo.Context) (string, error)
	Authenticate(ctx echo.Context) (*UserData, error)
}

const (
	sessionKey     = "query-adventure-auth"
	sessionUserKey = "user"
	sessionCtxKey  = "user"
)

type Middleware struct {
	authn Authenticator
}

func NewMiddleware(authn Authenticator) *Middleware {
	return &Middleware{authn: authn}
}

func UserSessionMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		sess, err := session.Get(sessionKey, ctx)
		if err != nil {
			return fmt.Errorf("error getting session: %w", err)
		}
		user, ok := sess.Values[sessionUserKey].(UserData)
		if !ok {
			return next(ctx)
		}
		ctx.Set(sessionCtxKey, &user)
		return next(ctx)
	}
}

func RequireUser() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user := c.Get(sessionCtxKey)
			if user == nil {
				return echo.NewHTTPError(http.StatusUnauthorized)
			}
			return next(c)
		}
	}
}

func User(e echo.Context) *UserData {
	return e.Get(sessionCtxKey).(*UserData)
}

func MustUser(e echo.Context) *UserData {
	user := User(e)
	if user == nil {
		panic("no user in context")
	}
	return user
}

func (am *Middleware) HandleSignIn(e echo.Context) error {
	redirectURL, err := am.authn.MaybeRedirect(e)
	if err != nil {
		return fmt.Errorf("error in %T.MaybeRedirect: %w", am.authn, err)
	}
	if redirectURL != "" {
		if acceptHdr := e.Request().Header.Get("Accept"); strings.Contains(acceptHdr, "application/json") {
			return e.JSON(http.StatusUnauthorized, map[string]string{
				"redirect": redirectURL,
			})
		}
		return e.Redirect(http.StatusTemporaryRedirect, redirectURL)
	}

	return am.completeSignIn(e)
}

func (am *Middleware) HandleRedirect(e echo.Context) error {
	return am.completeSignIn(e)
}

func (am *Middleware) completeSignIn(e echo.Context) error {
	user, err := am.authn.Authenticate(e)
	if err != nil {
		return fmt.Errorf("error in %T.Authenticate: %w", am.authn, err)
	}
	if user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}

	err = setUserSession(e, *user)
	if err != nil {
		return fmt.Errorf("error setting user session: %w", err)
	}

	return e.JSON(http.StatusOK, user)
}

func setUserSession(e echo.Context, u UserData) error {
	sess, err := session.Get(sessionKey, e)
	if err != nil {
		return fmt.Errorf("error getting session: %w", err)
	}
	sess.Values[sessionUserKey] = u
	return sess.Save(e.Request(), e.Response())
}
