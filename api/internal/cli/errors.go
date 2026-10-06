package cli

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

// Exit codes are a contract agents script against.
const (
	exitOK       = 0
	exitFailure  = 1
	exitUsage    = 2
	exitNotFound = 3
	exitAuth     = 4
	exitConflict = 5
)

type exitError struct {
	code int
	msg  string
}

func (e *exitError) Error() string { return e.msg }

func usageErr(format string, args ...any) error {
	return &exitError{code: exitUsage, msg: fmt.Sprintf(format, args...)}
}

func exitCode(err error) int {
	var exit *exitError
	var api *apiError
	var transport *transportError
	switch {
	case errors.As(err, &exit):
		return exit.code
	case errors.As(err, &transport):
		return exitFailure
	case errors.As(err, &api):
		switch api.Status {
		case http.StatusBadRequest:
			return exitUsage
		case http.StatusUnauthorized, http.StatusForbidden:
			return exitAuth
		case http.StatusNotFound:
			return exitNotFound
		case http.StatusConflict:
			return exitConflict
		}
		return exitFailure
	default:
		// cobra's own errors: bad flags, unknown commands.
		return exitUsage
	}
}

func (a *app) hint(cmd *cobra.Command, err error, code int) string {
	var api *apiError
	var transport *transportError
	switch {
	case errors.As(err, &transport):
		return "check the API is up and SHAPEHILL_API_URL is right (using " + a.apiURL + ")"
	case errors.As(err, &api) && api.Status == http.StatusUnauthorized:
		return "SHAPEHILL_TOKEN was rejected; it may be revoked. Create a new one at " + a.envOr("SHAPEHILL_WEB_URL", defaultWeb) + "/app/tokens"
	case code == exitUsage && cmd != nil && !strings.Contains(err.Error(), "usage:"):
		return fmt.Sprintf("run %s for usage", accent(cmd.CommandPath()+" --help"))
	}
	return ""
}
