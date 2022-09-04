package rest

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	"query-adventure/auth"
	"query-adventure/cfg"
	"query-adventure/data"
	"query-adventure/db"
	"query-adventure/ui"
)

type API struct {
	e    *echo.Echo
	g    *cfg.Globals
	cb   *db.CBDatabase
	ds   data.Datasets
	auth auth.Authenticator
	am   *auth.Middleware
}

func NewAPI(g *cfg.Globals, cb *db.CBDatabase, ds data.Datasets, authn auth.Authenticator) *API {
	a := &API{
		e:    echo.New(),
		g:    g,
		cb:   cb,
		ds:   ds,
		auth: authn,
		am:   auth.NewMiddleware(authn),
	}
	a.e.Logger.SetLevel(log.DEBUG)
	a.e.Use(middleware.Logger())
	a.e.Use(middleware.Recover())
	a.e.Use(session.Middleware(sessions.NewCookieStore([]byte(g.SessionKey))))
	a.e.Use(auth.UserSessionMiddleware)
	a.registerRoutes()
	return a
}

func (a *API) Start(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		err := a.e.Shutdown(ctx)
		if err != nil {
			a.e.Logger.Warnf("shutdown: %v", err)
		}
	}()
	addr := net.JoinHostPort("0.0.0.0", strconv.Itoa(a.g.HTTPPort))
	a.e.Logger.Infof("Starting on http://%s", addr)
	return a.e.Start(addr)
}

func (a *API) registerRoutes() {
	a.e.GET("/api/me", a.handleMe, auth.RequireUser())

	a.e.GET("/api/datasets", a.handleGetDatasets, auth.RequireUser())
	a.e.POST("/api/dataset/:ds/query", a.handleQuery, auth.RequireUser())
	a.e.POST("/api/dataset/:ds/:query/submitAnswer", a.handleSubmitAnswer, auth.RequireUser())

	a.e.GET("/api/signIn", a.am.HandleSignIn)
	a.e.POST("/api/signIn", a.am.HandleSignIn)
	a.e.GET("/api/signIn/redirect", a.am.HandleRedirect)

	if ui.EmbeddedUI != nil {
		a.e.StaticFS("/", ui.EmbeddedUI)
	}
}

func (a *API) handleQuery(c echo.Context) error {
	ds, err := a.getDS(c)
	if err != nil {
		return err
	}
	var body struct {
		Statement string `json:"statement" form:"statement"`
	}
	err = c.Bind(&body)
	if err != nil {
		return err
	}

	err = checkAndSetRateLimit(c, rlQuery, a.g)
	if err != nil {
		return err
	}

	res, err := a.cb.ExecuteQuery(c.Request().Context(), ds.Keyspace, body.Statement)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to execute query: %v", err))
	}

	return c.JSON(http.StatusOK, map[string]any{
		"rows": res,
	})
}

func (a *API) handleSubmitAnswer(c echo.Context) error {
	ds, err := a.getDS(c)
	if err != nil {
		return err
	}
	queryIdx, err := strconv.Atoi(c.Param("query"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid query: %w", err)
	}
	if queryIdx < 0 || queryIdx >= len(ds.Queries) {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid query %d", queryIdx)
	}
	var body struct {
		Statement string `json:"statement" form:"statement"`
	}
	err = c.Bind(&body)
	if err != nil {
		return err
	}
	query := ds.Queries[queryIdx]

	err = checkAndSetRateLimit(c, rlCheck, a.g)
	if err != nil {
		return err
	}

	err = a.cb.ExecuteAndVerifyQuery(c.Request().Context(), ds.Keyspace, query.Query, body.Statement)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK) // FIXME set points
}

func (a *API) getDS(c echo.Context) (data.Dataset, error) {
	dsStr := c.Param("ds")
	ds, err := strconv.Atoi(dsStr)
	if err != nil {
		return data.Dataset{}, echo.NewHTTPError(http.StatusBadRequest, "invalid ds: %w", err)
	}
	if ds < 0 || ds >= len(a.ds) {
		return data.Dataset{}, echo.NewHTTPError(http.StatusBadRequest, "invalid ds %d", ds)
	}
	return a.ds[ds], nil
}

func (a *API) handleGetDatasets(c echo.Context) error {
	return c.JSON(http.StatusOK, a.ds.FilterQueries())
}

func (a *API) handleMe(c echo.Context) error {
	user := auth.MustUser(c)
	return c.JSON(http.StatusOK, user)
}
