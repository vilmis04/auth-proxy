package auth

import (
	"encoding/json"
	"net/http"

	"github.com/vilmis04/auth-proxy/internal/accessToken"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) getIsAuthorized(token string) bool {
	user, err := accessToken.Validate(token)
	if err != nil {
		return false
	}
	if user == nil || *user == "" {
		return false
	}

	return true
}

func (s *Service) signUp(request *http.Request) (*string, error) {
	var body signUpRequest
	err := json.NewDecoder(request.Body).Decode(&body)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
