package api

import (
	"net/url"
	"time"

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
