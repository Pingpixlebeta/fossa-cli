// Package fossa provides a high-level interface to the FOSSA API (by default,
// located at https://app.fossa.io).
package fossa

import (
	"net/url"

	"github.com/fossas/fossa-cli/api"
	"github.com/fossas/fossa-cli/errors"
)

var (
	serverURL *url.URL
	apiKey    string
)

// SetEndpoint sets the URL of the FOSSA backend (e.g. for on-premises or local
// development instances), returning an error if the URL is invalid.
func SetEndpoint(endpoint string) error {
	u, err := url.Parse(endpoint)
	if err != nil {
		return err
	}
	serverURL = u
	return nil
}

// SetAPIKey sets the package's API key, returning ErrNoAPIKey if the key is an
// empty string.
func SetAPIKey(key string) *errors.Error {
	// TODO: we should check that the API key is valid and that permissions are
	// correct for this project _before_ doing anything else. Invalid API keys
	// should be an error distinct from missing API keys.
	if key == "" {
		return &errors.Error{
			Code:    "E_MISSING_API_KEY",
			Type:    errors.UserInput,
			Message: "missing FOSSA API key",
			Troubleshooting: `

Try providing your API key by setting the $FOSSA_API_KEY environment variable.
For example, try running:

	FOSSA_API_KEY=your-api-key-here %s

You can create an API key for your FOSSA account at:

	<https://app.fossa.io/account/settings/integrations/api_tokens>

`,
		}
	}
	apiKey = key
	return nil
}

// Get makes an authenticated GET request to a FOSSA API endpoint.
func Get(endpoint string) (res string, statusCode int, err error) {
	u, err := serverURL.Parse(endpoint)
	if err != nil {
		return "", 0, err
	}
	return api.Get(u, apiKey, nil)
}

// GetJSON makes an authenticated JSON GET request to a FOSSA API endpoint.
func GetJSON(endpoint string, v interface{}) (statusCode int, err error) {
	u, err := serverURL.Parse(endpoint)
	if err != nil {
		return 0, err
	}
	return api.GetJSON(u, apiKey, nil, v)
}

// Post makes an authenticated POST request to a FOSSA API endpoint.
// TODO: maybe `body` should be an `io.Reader` instead.
func Post(endpoint string, body []byte) (res string, statusCode int, err error) {
	u, err := serverURL.Parse(endpoint)
	if err != nil {
		return "", 0, err
	}
	return api.Post(u, apiKey, body)
}
