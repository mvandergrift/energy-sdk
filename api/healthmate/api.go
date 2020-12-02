package healthmate

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload" // autoload configuration

	"github.com/jinzhu/gorm"
	"golang.org/x/oauth2"

	"github.com/mvandergrift/energy-sdk/api"
	"github.com/mvandergrift/energy-sdk/auth"
	"github.com/mvandergrift/energy-sdk/driver"
)

const debug = false

type Client api.Client // allow receivers on shared Client definition

var (
	// DefaultScopes is describes all available Health Mate scopes.
	// Needs to be comma separated for the Health Mate endpoint and slice of string
	// for Oauth2 package.
	DefaultScopes = []string{"user.info,user.metrics,user.activity,user.sleepevents"}

	// HealthMateState is a random state for generating auth code url to mitigate CSRF attacks.
	HealthMateState string
)

// NewClient instantiates a new client to interact with the Health Mate API. Please
// refer to the official Withings documentation to obtain the required parameters
func NewClient(clientID, clientSecret, redirectURL string) api.ApiClient {
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

// GetAuthCodeURL obtains the user authentication URL
func (c Client) GetAuthCodeURL() string {
	return c.OAuth2Config.AuthCodeURL(HealthMateState)
}

// GetAccessToken obtains the access token for the authenticated user
func (c Client) GetAccessToken(code string) (token *oauth2.Token, err error) {
	ctx, cancel := c.getContext()
	defer cancel()

	token, err = c.OAuth2Config.Exchange(ctx, code)

	return token, err
}

func (hc Client) ProcessRequest(payload url.Values, v interface{}) error {
	client, err := newHTTPClient(hc)
	if err != nil {
		return fmt.Errorf("NewHTTPClient %w", err)
	}

	resp, err := client.PostForm("https://wbsapi.withings.net/v2/measure", payload)
	if err != nil {
		return fmt.Errorf("PostForm %w", err)
	}

	defer resp.Request.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ioutil.ReadAll %w", err)
	}

	// if *debugFlag {
	// 	err = ioutil.WriteFile("debug.json", body, 0644)
	// 	if err != nil {
	// 		return fmt.Errorf("WriteFile (debug) %w", err)
	// 	}
	// }

	err = json.Unmarshal(body, v)
	if err != nil {
		return fmt.Errorf("Unmarshal %w", err)
	}

	return nil
}

func init() {
	var err error

	HealthMateState, err = randState()
	if err != nil {
		panic(err)
	}
}

// randState randomly generates a base64 encoded string of 10 bytes
func randState() (state string, err error) {
	buffer := make([]byte, 10)
	_, err = rand.Read(buffer)
	return base64.URLEncoding.EncodeToString(buffer), err
}

/*
Returns a new HTTPClient based on the Healthmate OAuth2 client & configurations. Uses
the cachedTokenPath to load and store the access token needed for api authentication
*/
func newHTTPClient(c Client) (*http.Client, error) {
	db := getDb()
	token, err := auth.LoadToken(db)
	if err != nil {
		return nil, err
	}

	tokenSource := auth.RefreshToken(token, c.OAuth2Config, db)
	client := oauth2.NewClient(context.Background(), *tokenSource)
	return client, nil
}

// getContext creates a context
func (c *Client) getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), c.Timeout)
}

func getDb() *gorm.DB {
	db, err := driver.OpenCn(os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PWD"), os.Getenv("DB_DATABASE"), debug)
	if err != nil {
		panic(fmt.Sprintf("Cannot connect to token DB: %v", err))
	}

	return db
}

// setScopes is a variadic function that sets the required scopes
// for your application for authentication with users.
// Default behaviour is to use all scopes available.
// func (c *Client) setScopes(scopes ...string) {
// 	formatted := strings.Join(scopes, ",")

// 	c.OAuth2Config.Scopes = []string{formatted}
// }
