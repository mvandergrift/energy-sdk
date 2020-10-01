package healthmate

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

var (
	// DefaultScopes is describes all available Health Mate scopes.
	// Needs to be comma separated for the Health Mate endpoint and slice of string
	// for Oauth2 package.
	DefaultScopes = []string{"user.info,user.metrics,user.activity,user.sleepevents"}

	// HealthMateState is a random state for generating auth code url to mitigate CSRF attacks.
	HealthMateState string
)

func init() {
	var err error

	HealthMateState, err = randState()
	if err != nil {
		panic(err)
	}
}

// randState randomly generates a base64 encoded string of
// 10 bytes
func randState() (state string, err error) {
	buffer := make([]byte, 10)
	_, err = rand.Read(buffer)
	return base64.URLEncoding.EncodeToString(buffer), err
}

// Client provides methods to to interact with the Health Mate API
type Client struct {
	OAuth2Config *oauth2.Config
	Timeout      time.Duration
}

// NewClient instantiates a new client to interact with the Health Mate API. Please
// refer to the official Withings documentation to obtain the required parameters
func NewClient(clientID, clientSecret, redirectURL string) Client {
	return Client{
		OAuth2Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint:     HealthMateEndpoint,
			RedirectURL:  redirectURL,
			Scopes:       DefaultScopes,
		},
		Timeout: 5 * time.Second,
	}
}

// SetScopes is a variadic function that sets the required scopes
// for your application for authentication with users.
// Default behaviour is to use all scopes available.
func (c *Client) SetScopes(scopes ...string) {
	formatted := strings.Join(scopes, ",")

	c.OAuth2Config.Scopes = []string{formatted}
}

// getContext creates a context
func (c *Client) getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), c.Timeout)
}

// GetAuthCodeURL obtains the user authentication URL
func (c *Client) GetAuthCodeURL() string {
	return c.OAuth2Config.AuthCodeURL(HealthMateState)
}

// GetAccessToken obtains the access token for the authenticated user
func (c *Client) GetAccessToken(code string) (token *oauth2.Token, err error) {
	ctx, cancel := c.getContext()
	defer cancel()

	token, err = c.OAuth2Config.Exchange(ctx, code)

	return token, err
}
