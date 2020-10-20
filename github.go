package github

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type AccessToken struct {
	Token               string            `json:"token"`
	ExpireAt            string            `json:"expires_at"`
	Permissions         map[string]string `json:"permissions"`
	RepositorySelection string            `json:"repository_selection"`
}

func getJWT(key []byte, appID int64) (string, error) {
	signKey, err := jwt.ParseRSAPrivateKeyFromPEM(key)
	if err != nil {
		return "", err
	}

	t := jwt.New(jwt.GetSigningMethod("RS256"))
	claims := make(jwt.MapClaims)
	claims["iat"] = time.Now().Unix()
	claims["iss"] = appID
	claims["exp"] = time.Now().Add(time.Minute * 10).Unix()
	t.Claims = claims
	tokenString, err := t.SignedString(signKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func getInstallationToken(key []byte, appID int64, installationID string) (AccessToken, error) {
	url := fmt.Sprintf("https://api.github.com/app/installations/%s/access_tokens", installationID)
	tokenString, err := getJWT(key, appID)
	if err != nil {
		return AccessToken{}, err
	}

	c := http.Client{}
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return AccessToken{}, err
	}
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	req.Header.Add("Authorization", "Bearer "+tokenString)
	res, err := c.Do(req)
	if err != nil {
		return AccessToken{}, err
	}

	defer res.Body.Close()

	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return AccessToken{}, err
	}
	bodyString := string(bodyBytes)
	accesstToken := AccessToken{}
	err = json.Unmarshal([]byte(bodyString), &accesstToken)
	if err != nil {
		return AccessToken{}, err
	}
	return accesstToken, nil

}
