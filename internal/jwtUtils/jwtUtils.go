package jwtUtils

import (
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

var key []byte = []byte(os.Getenv("jwt_key"))

func CreateJWT(username string) (*string, error) {
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

func ValidateJWT(jwtCookie string) (*string, error) {
	token, err := jwt.Parse(jwtCookie, func(t *jwt.Token) (interface{}, error) {
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
