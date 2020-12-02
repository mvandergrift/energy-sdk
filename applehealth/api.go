package applehealth

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"golang.org/x/oauth2"
)

var (
	DefaultScopes    = []string{"user.info, user.metrics, user.activity"}
	AppleHealthState string
)

// HealthMateEndpoint is the endpoints for Withings Health Mate
var AppleHealthEndpoint = oauth2.Endpoint{
	AuthURL:  "https://account.apple.com/oauth2_user/authorize2",
	TokenURL: "https://account.apple.com/oauth2/token",
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

type Client struct {
	OAuth2Config *oauth2.Config
	Timeout      time.Duration
}

func NewClient(clientID string, clientSecret string, redirectURL string) Client {
	return Client{
		OAuth2Config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint:     AppleHealthEndpoint,
			RedirectURL:  redirectURL,
			Scopes:       DefaultScopes,
		},
		Timeout: 5 * time.Second,
	}
}
