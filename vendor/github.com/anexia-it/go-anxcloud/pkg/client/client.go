// Package client provides a client for interacting with the anxcloud API.
package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"time"
)

const (
	// TokenEnvName is the name of the environment variable that should contain the API token.
	TokenEnvName = "ANEXIA_TOKEN" //nolint:gosec // This is a name, not a secret.
	// LocationEnvName is the name of the environment variable that should contain the location of VMs to manage.
	LocationEnvName = "ANEXIA_LOCATION_ID"
	// VLANEnvName is the name of the environment variable that should contain the VLAN of VMs to manage.
	VLANEnvName = "ANEXIA_VLAN_ID"
	// IntegrationTestEnvName is the name of the environment variable that enables integration tests if present.
	IntegrationTestEnvName = "ANEXIA_INTEGRATION_TESTS_ON"
	// DefaultBaseURL is the default base URL used for requests.
	DefaultBaseURL = "https://engine.anexia-it.com"
	// DefaultRequestTimeout is a suggested timeout for API calls.
	DefaultRequestTimeout = 10 * time.Second
)

// ErrEnvMissing indicates an environment variable is missing.
var ErrEnvMissing = errors.New("environment variable missing")

// Client interacts with the anxcloud API.
type Client interface {
	// Do fires a given http.Request against the API.
	// This method behaves as http.Client.Do, but signs the request prior to sending it out
	// and returns an error is the response status is not OK.
	Do(req *http.Request) (*http.Response, error)
	BaseURL() string
}

// ResponseError is a response from the API that indicates an error.
type ResponseError struct {
	Request   *http.Request  `json:"-"`
	Response  *http.Response `json:"-"`
	ErrorData struct {
		Code       int               `json:"code"`
		Message    string            `json:"message"`
		Validation map[string]string `json:"validation"`
	} `json:"error"`
	Debug struct {
		Source string `json:"source"`
	} `json:"debug"`
}

func (r ResponseError) Error() string {
	return fmt.Sprintf("received error from api: %+v", r.ErrorData)
}

func handleRequest(c *http.Client, req *http.Request) (*http.Response, error) {
	reqBytes, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not dump req: %+v", err)
	} else {
		fmt.Println(string(reqBytes))
	}

	response, err := c.Do(req)
	if err == nil && response.StatusCode != http.StatusOK {
		errResponse := ResponseError{Request: req, Response: response}
		if decodeErr := json.NewDecoder(response.Body).Decode(&errResponse); decodeErr != nil {
			return response, fmt.Errorf("could not decode error response: %w", decodeErr)
		}

		return response, &errResponse
	}

	return response, err
}

type optionSet struct {
	httpClient *http.Client
	token      string
}

// Option is a optional parameter for the New method.
type Option func(o *optionSet) error

// AuthFromEnv uses any known environment variables to create a client.
func AuthFromEnv(unset bool) Option {
	return TokenFromEnv(unset)
}

// TokenFromString uses the given API auth token.
func TokenFromString(token string) Option {
	return func(o *optionSet) error {
		o.token = token

		return nil
	}
}

// TokenFromEnv fetches the API auth token from environment variables.
func TokenFromEnv(unset bool) Option {
	return func(o *optionSet) error {
		token, tokenPresent := os.LookupEnv(TokenEnvName)
		if !tokenPresent {
			return fmt.Errorf("%w: %s", ErrEnvMissing, TokenEnvName)
		}
		o.token = token
		if unset {
			if err := os.Unsetenv(TokenEnvName); err != nil {
				return fmt.Errorf("could not unset %s: %w", TokenEnvName, err)
			}
		}

		return nil
	}
}

// HTTPClient lets the client use the given http.Client.
func HTTPClient(c *http.Client) Option {
	return func(o *optionSet) error {
		o.httpClient = c

		return nil
	}
}

// ErrConfiguration is raised when the given configuration is insufficient or erroneous.
var ErrConfiguration = errors.New("could not configure client")

// New creates a new client with the given options.
//
// The options need to contain a method of authentication with the API. If you are
// unsure what to use pass AuthFromEnv.
func New(options ...Option) (Client, error) {
	optionSet := optionSet{}
	for _, option := range options {
		if err := option(&optionSet); err != nil {
			return nil, err
		}
	}
	if optionSet.httpClient == nil {
		optionSet.httpClient = http.DefaultClient
	}

	if optionSet.token != "" {
		return &tokenClient{optionSet.token, optionSet.httpClient}, nil
	}

	return nil, fmt.Errorf("%w: token not set", ErrConfiguration)
}
