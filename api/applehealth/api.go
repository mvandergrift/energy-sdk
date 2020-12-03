package applehealth

import (
	"crypto/rand"
	"encoding/base64"
	"net/url"
	"time"

	"golang.org/x/oauth2"

	"github.com/mvandergrift/energy-sdk/api"
)

type Client api.Client

var (
	DefaultScopes    = []string{"user.info, user.metrics, user.activity"}
	AppleHealthState string
)

// HealthMateEndpoint is the endpoints for Withings Health Mate
var AppleHealthEndpoint = oauth2.Endpoint{
	AuthURL:  "https://account.apple.com/oauth2_user/authorize2",
	TokenURL: "https://account.apple.com/oauth2/token",
}

func NewClient(clientID string, clientSecret string, redirectURL string) api.ApiClient {
	x := &Client{
		OAuth2Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint:     AppleHealthEndpoint,
			RedirectURL:  redirectURL,
			Scopes:       DefaultScopes,
		},
		Timeout: 5 * time.Second,
	}

	return x
}

// GetAuthCodeURL obtains the user authentication URL
func (c *Client) GetAuthCodeURL() string {
	return "Not Implemented"
}

// GetAccessToken obtains the access token for the authenticated user
func (c *Client) GetAccessToken(code string) (token *oauth2.Token, err error) {
	return nil, nil
}

func (hc Client) ProcessRequest(payload url.Values, v interface{}) error {
	return nil
}

func init() {
	var err error
	AppleHealthState, err = randState()
	if err != nil {
		panic(err)
	}
}

func randState() (string, error) {
	buffer := make([]byte, 10)
	_, err := rand.Read(buffer)
	return base64.URLEncoding.EncodeToString(buffer), err
}
