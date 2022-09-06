package rest

import (
	"fmt"

	"query-adventure/auth"
	"query-adventure/data"
	"query-adventure/rest/ratelimit"

	"github.com/labstack/echo/v4"
)

const rlQuery = ratelimit.Key("query")
const rlCheck = ratelimit.Key("check")

func (a *API) casQueryLimit(e echo.Context) error {
	user := auth.MustUser(e)
	return a.rl.CheckAndSet(rlQuery, user.Email)
}

func (a *API) casCheckLimit(e echo.Context, query data.Query) error {
	user := auth.MustUser(e)
	team, err := a.mCB.GetTeamForUser(e.Request().Context(), user.Email)
	if err != nil {
		return fmt.Errorf("failed to get team for user %q: %w", user.Email, err)
	}

	return a.rl.CheckAndSet(rlCheck, query.ID+"::"+team.ID)
}
