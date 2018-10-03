// Package setup implements initialization for all application packages.
package setup

import (
	"github.com/urfave/cli"

	"github.com/fossas/fossa-cli/api/fossa"
	"github.com/fossas/fossa-cli/cmd/fossa/display"
	"github.com/fossas/fossa-cli/config"
	"github.com/fossas/fossa-cli/errors"
)

// SetContext initializes all application-level packages.
func SetContext(ctx *cli.Context) error {
	// Set up configuration.
	err := config.SetContext(ctx)
	if err != nil {
		return err
	}

	// Set up logging.
	display.SetInteractive(config.Interactive())
	display.SetDebug(config.Debug())

	// Set up API.
	err = fossa.SetEndpoint(config.Endpoint())
	if err != nil {
		return err
	}

	return nil
}

// SetAPIKey sets the API key for the API package.
func SetAPIKey(key string) error {
	if key == "" {
		return errors.Error{
			Code:    "E_MISSING_API_KEY",
			Type:    errors.UserInput,
			Message: "missing FOSSA API key",
		}
	}
	fossa.SetAPIKey(key)
	return nil
}
