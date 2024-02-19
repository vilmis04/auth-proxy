package accessToken

import (
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

const ACCESS_TOKEN = "access_token"

var key []byte = []byte(os.Getenv("jwt_key"))

func Create(username string) (*string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub": username,
		})
	signedToken, err := token.SignedString(key)
	if err != nil {
		return nil, err
	}

	return &signedToken, nil
}

func Validate(tokenCookie string) (*string, error) {
	token, err := jwt.Parse(tokenCookie, func(t *jwt.Token) (interface{}, error) {
		return key, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("jwt token is not valid")
	}
	user, err := token.Claims.GetSubject()
	if err != nil {
		return nil, err
	}

	return &user, nil
}
