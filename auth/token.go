package auth

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"golang.org/x/oauth2"
)

// Refreshes token and compares to previous value. Saves to disk if it has changed
func RefreshToken(token *oauth2.Token, conf *oauth2.Config, filePath string) *oauth2.TokenSource {
	tokenSource := conf.TokenSource(context.Background(), token)
	newToken, err := tokenSource.Token()
	if err != nil {
		log.Fatalln("Fetch Token", err)
	}

	if newToken.AccessToken != token.AccessToken {
		// save new access &refresh token
		err := SaveToken(newToken, filePath)
		if err != nil {
			log.Fatalln("SaveToken:", err)
		}
	}

	return &tokenSource
}

func SaveToken(token *oauth2.Token, filePath string) error {
	b, _ := json.Marshal(&token)
	f, _ := os.Create(filePath)
	defer f.Close()
	_, err := f.WriteString(string(b))
	return err
}

func LoadToken(filePath string) (*oauth2.Token, error) {
	buffer, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	token := new(oauth2.Token)
	err = json.Unmarshal(buffer, &token)
	return token, err
}
