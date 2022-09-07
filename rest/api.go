package rest

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"golang.org/x/exp/slices"

	"query-adventure/auth"
	"query-adventure/cfg"
	"query-adventure/data"
	"query-adventure/db"
	"query-adventure/rest/ratelimit"
	"query-adventure/ui"
)

type API struct {
	e    *echo.Echo
	g    *cfg.Globals
	qCB  *db.QueryConnection
	mCB  *db.ManagementConnection
	ds   data.Datasets
	auth auth.Authenticator
	am   *auth.Middleware
	rl   *ratelimit.RateLimiter
}

func NewAPI(g *cfg.Globals, qCB *db.QueryConnection, mCB *db.ManagementConnection, ds data.Datasets, authn auth.Authenticator) *API {
	a := &API{
		e:    echo.New(),
		g:    g,
		qCB:  qCB,
		mCB:  mCB,
		ds:   ds,
		auth: authn,
		am:   auth.NewMiddleware(authn),
		rl: ratelimit.NewRateLimiter(map[ratelimit.Key]time.Duration{
			rlQuery: g.RateLimits[string(rlQuery)],
			rlCheck: g.RateLimits[string(rlCheck)],
		}),
	}
	a.e.Logger.SetLevel(log.DEBUG)
	a.e.HTTPErrorHandler = a.errorHandler
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
	a.e.POST("/api/dataset/:ds/:query/useHint", a.handleUseHint, auth.RequireUser())

	a.e.GET("/api/scoreboard", a.handleScoreboard, auth.RequireUser())
	a.e.GET("/api/completedChallenges", a.handleCompletedChallenges, auth.RequireUser())
	a.e.GET("/api/teams", a.handleTeams, auth.RequireUser())

	a.e.GET("/api/signIn", a.am.HandleSignIn)
	a.e.POST("/api/signIn", a.am.HandleSignIn)
	a.e.GET("/api/signIn/redirect", a.am.HandleRedirect)

	if ui.EmbeddedUI != nil {
		a.e.StaticFS("/", ui.EmbeddedUI)
	}
}

func (a *API) handleQuery(c echo.Context) error {
	ds, ok := a.ds.DatasetByID(c.Param("ds"))
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, "no such dataset")
	}

	var body struct {
		Statement string `json:"statement" form:"statement"`
	}
	err := c.Bind(&body)
	if err != nil {
		return err
	}

	err = a.casQueryLimit(c) // TODO: index creation should be different
	if err != nil {
		return err
	}

	res, err := a.qCB.ExecuteQuery(c.Request().Context(), ds.Keyspace, body.Statement)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to execute query: %v", err))
	}

	return c.JSON(http.StatusOK, map[string]any{
		"rows": res,
	})
}

type CorrectAnswerResponse struct {
	OK     bool    `json:"ok"`
	Points float64 `json:"points"`
}

func (a *API) handleSubmitAnswer(c echo.Context) error {
	ds, ok := a.ds.DatasetByID(c.Param("ds"))
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, "no such dataset")
	}
	query, ok := ds.QueryByID(c.Param("query"))
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, "query not found")
	}

	var body struct {
		Statement string `json:"statement" form:"statement"`
	}
	err := c.Bind(&body)
	if err != nil {
		return err
	}

	user := auth.MustUser(c)
	team, err := a.mCB.GetTeamForUser(c.Request().Context(), user.Email)
	if err != nil {
		return fmt.Errorf("failed to get team: %w", err)
	}

	hints, err := a.mCB.GetUsedHints(c.Request().Context(), ds.ID, query.ID, team.ID)
	if err != nil {
		return fmt.Errorf("failed to get hints total: %w", err)
	}

	err = a.casCheckLimit(c, query)
	if err != nil {
		return err
	}

	err = a.qCB.ExecuteAndVerifyQuery(c.Request().Context(), ds.Keyspace, query.Query, body.Statement)
	if err != nil {
		return err
	}

	cc, err := a.mCB.CompleteChallenge(c.Request().Context(), a.g, ds, query, team, user.Email, body.Statement, hints)
	if err != nil {
		return fmt.Errorf("failed to mark challenge %s.%s as complete: %w", ds.ID, query.ID, err)
	}

	return c.JSON(http.StatusOK, CorrectAnswerResponse{
		OK:     true,
		Points: cc.FinalPoints,
	})
}

type apiDataset struct {
	data.Dataset
	Queries []apiQuery `json:"queries"`
}

type apiQuery struct {
	data.Query
	Complete bool `json:"complete"`
	NumHints int  `json:"numHints"`
}

func (a *API) handleGetDatasets(c echo.Context) error {
	rawData := a.ds
	user := auth.MustUser(c)
	team, err := a.mCB.GetTeamForUser(c.Request().Context(), user.Email)
	if err != nil {
		return fmt.Errorf("failed to get user team: %w", err)
	}
	complete, err := a.mCB.GetTeamCompleteChallenges(c.Request().Context(), team)
	if err != nil {
		return fmt.Errorf("failed to find complete challenges: %w", err)
	}
	result := make([]apiDataset, 0, len(rawData))
	for _, d := range rawData {
		ds := apiDataset{
			Dataset: d,
			Queries: make([]apiQuery, 0, len(d.Queries)),
		}
		for _, q := range d.Queries {
			usedHints, err := a.mCB.GetUsedHints(c.Request().Context(), d.ID, q.ID, team.ID)
			if err != nil {
				return fmt.Errorf("failed to get used hints for %s.%s: %w", ds.ID, q.ID, err)
			}
			ds.Queries = append(ds.Queries, makeAPIQuery(d, q, usedHints, complete))
		}
		result = append(result, ds)
	}
	return c.JSON(http.StatusOK, result)
}

func makeAPIQuery(ds data.Dataset, q data.Query, usedHints uint, complete map[string][]string) apiQuery {
	return apiQuery{
		Query:    q.FilterForPublic(usedHints),
		NumHints: len(q.Hints),
		Complete: slices.Contains(complete[ds.ID], q.ID),
	}
}

func (a *API) handleUseHint(c echo.Context) error {
	ds, ok := a.ds.DatasetByID(c.Param("ds"))
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, "no such dataset")
	}
	query, ok := ds.QueryByID(c.Param("query"))
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, "query not found")
	}

	user := auth.MustUser(c)
	team, err := a.mCB.GetTeamForUser(c.Request().Context(), user.Email)
	if err != nil {
		return fmt.Errorf("failed to get team: %w", err)
	}

	curr, used, err := a.mCB.UseHint(c.Request().Context(), ds.ID, query.ID, team.ID, len(query.Hints))
	if err != nil {
		return fmt.Errorf("failed to use hint: %w", err)
	}
	if !used {
		return echo.NewHTTPError(http.StatusBadRequest, "all hints already used")
	}

	complete, err := a.mCB.GetTeamCompleteChallenges(c.Request().Context(), team)
	if err != nil {
		return fmt.Errorf("failed to find complete challenges: %w", err)
	}

	return c.JSON(http.StatusOK, makeAPIQuery(ds, query, curr, complete))
}

func (a *API) handleMe(c echo.Context) error {
	user := auth.MustUser(c)
	return c.JSON(http.StatusOK, user)
}

func (a *API) errorHandler(err error, c echo.Context) {
	var httpErr *echo.HTTPError
	if errors.As(err, &httpErr) {
		a.e.DefaultHTTPErrorHandler(httpErr, c)
		return
	}
	a.e.DefaultHTTPErrorHandler(err, c)
}

func (a *API) handleScoreboard(c echo.Context) error {
	res, err := a.mCB.GetTeamScores(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}

func (a *API) handleCompletedChallenges(c echo.Context) error {
	res, err := a.mCB.GetAllTeamCompleteChallenges(c.Request().Context(), a.ds)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}

func (a *API) handleTeams(c echo.Context) error {
	res, err := a.mCB.GetAllTeams(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}
