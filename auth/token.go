package auth

import (
	"context"
	"encoding/json"
	"log"

	"github.com/jinzhu/gorm"
	"golang.org/x/oauth2"

	"github.com/mvandergrift/energy-sdk/model"
)

const userId = 1

// Refreshes token and compares to previous value. Saves to disk if it has changed
func RefreshToken(token *oauth2.Token, conf *oauth2.Config, cn *gorm.DB) *oauth2.TokenSource {
	tokenSource := conf.TokenSource(context.Background(), token)
	newToken, err := tokenSource.Token()
	if err != nil {
		log.Fatalln("Fetch Token", err)
	}

	if newToken.AccessToken != token.AccessToken {
		// save new access &refresh token
		err := SaveToken(newToken, cn)
		if err != nil {
			log.Fatalln("SaveToken:", err)
		}
	}

	return &tokenSource
}

func SaveToken(token *oauth2.Token, db *gorm.DB) error {
	b, _ := json.Marshal(&token)
	return (db.Model(&model.User{}).Where("id = ?", userId).Update("withings_token", string(b)).Error)
}

func LoadToken(db *gorm.DB) (*oauth2.Token, error) {
	var user model.User
	err := db.First(&user, "id = ?", userId).Error
	if err != nil {
		return nil, err
	}

	token := new(oauth2.Token)
	err = json.Unmarshal([]byte(user.WithingsToken), &token)
	return token, err
}
