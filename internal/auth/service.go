package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"

	"github.com/vilmis04/auth-proxy/internal/accessToken"
)

type Service struct {
	Repo
}

func NewService() *Service {
	return &Service{
		Repo: *NewRepo(),
	}
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

func (s *Service) validateSignUpRequest(body signUpRequest) error {
	if body.Password != body.RepeatPassword {
		return fmt.Errorf("passwords do not match")
	}

	names, err := s.Repo.GetUserList()
	if err != nil {
		return err
	}

	if slices.Contains(*names, body.Username) {
		return fmt.Errorf("username %v is already taken", body.Username)
	}

	return nil
}

func (s *Service) signUp(request *http.Request) (*string, error) {
	var body signUpRequest
	err := json.NewDecoder(request.Body).Decode(&body)
	if err != nil {
		return nil, err
	}

	err = s.validateSignUpRequest(body)
	if err != nil {
		return nil, err
	}

	err = s.Repo.createUser(body)
	if err != nil {
		return nil, err
	}

	token, err := accessToken.Create(body.Username)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (s *Service) checkUser(body *loginRequest) error {
	user, err := s.Repo.getUser(body.Username)
	if err != nil {
		return err
	}

	return ValidatePassword(body.Password, user.Password)
}

func (s *Service) login(request *http.Request) (*string, error) {
	var body loginRequest
	err := json.NewDecoder(request.Body).Decode(&body)
	if err != nil {
		return nil, err
	}

	err = s.checkUser(&body)
	if err != nil {
		return nil, err
	}

	token, err := accessToken.Create(body.Username)
	if err != nil {
		return nil, err
	}

	return token, nil
}
