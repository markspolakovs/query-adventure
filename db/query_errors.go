package db

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func errNotEnoughRows(expected, actual uint) error {
	return echo.NewHTTPError(http.StatusExpectationFailed, fmt.Sprintf("Your query did not return as many rows as it should have done (we expected %d, but only got %d).", expected, actual))
}

func errTooManyRows(expected uint, actual uint) error {
	return echo.NewHTTPError(http.StatusExpectationFailed, fmt.Sprintf("Your query returned too many rows (we expected %d, but got %d).", expected, actual))
}

func errMismatch(row uint, val any) error {
	return echo.NewHTTPError(http.StatusExpectationFailed, fmt.Sprintf("Your query gave an unexpected result on row %d: %+v", row, val))
}
