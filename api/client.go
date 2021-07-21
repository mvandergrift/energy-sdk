package api

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/mvandergrift/energy-sdk/driver"
	"golang.org/x/oauth2"
)

type Client struct {
	OAuth2Config *oauth2.Config
	Timeout      time.Duration
}

type ApiClient interface {
	GetAuthCodeURL() string
	GetAccessToken(code string) (token *oauth2.Token, err error)
	ProcessRequest(payload url.Values, v interface{}) error
}

// RandState is a stnadardized randomization and initialization function used to setup a new API module
func RandState() (string, error) {
	buffer := make([]byte, 10)
	_, err := rand.Read(buffer)
	return base64.URLEncoding.EncodeToString(buffer), err
}

// GetDb returns an initalized connection string based on the standardized envioronment configuration (hosted or local)
func GetDb() *gorm.DB {
	db, err := driver.OpenCn(os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PWD"), os.Getenv("DB_DATABASE"), false)
	if err != nil {
		panic(fmt.Sprintf("Cannot connect to token DB: %v", err))
	}

	return db
}
