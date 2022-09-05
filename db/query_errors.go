package db

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func mustMarshalJSON(row any) []byte {
	jv, _ := json.Marshal(row)
	return jv
}

func errNotEnoughRows(expected, actual uint, lastSeen, nextWanted any) error {
	return echo.NewHTTPError(http.StatusExpectationFailed, fmt.Sprintf(
		"Your query did not return as many rows as it should have done (we expected %d, but only got %d). The last row your query returned was %s, and the next we expected would have been %s.",
		expected, actual, mustMarshalJSON(lastSeen), mustMarshalJSON(nextWanted)))
}

func errTooManyRows(expected uint, actual uint, lastWanted, nextSeen any) error {
	return echo.NewHTTPError(http.StatusExpectationFailed, fmt.Sprintf(
		"Your query returned too many rows (we expected %d, but got %d). The last row we expected was %s, and the next one your query returned was %s.",
		expected, actual, mustMarshalJSON(lastWanted), mustMarshalJSON(nextSeen)))
}

func errMismatch(row uint, expected, actual any) error {
	return echo.NewHTTPError(http.StatusExpectationFailed, fmt.Sprintf(
		"Your query gave an unexpected result on row %d: we were expecting to see %s, but saw %s",
		row, mustMarshalJSON(expected), mustMarshalJSON(actual)))
}
