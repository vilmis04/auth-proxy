package auth

import (
	"encoding/json"
	"fmt"
	"log"
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

func (s *Service) getIsAuthenticated(token string) (*string, error) {
	user, err := accessToken.Validate(token)
	if err != nil {
		return nil, err
	}
	if user == nil || *user == "" {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
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

func (s *Service) signUp(request *http.Request) (token *string, serverErr error, clientErr error) {
	var body signUpRequest
	err := json.NewDecoder(request.Body).Decode(&body)
	if err != nil {
		return nil, fmt.Errorf("JSON err: %v", err), nil
	}

	err = s.validateSignUpRequest(body)
	if err != nil {
		return nil, nil, err
	}

	err = s.Repo.createUser(body)
	if err != nil {
		return nil, fmt.Errorf("user creation err: %v", err), nil
	}

	token, err = accessToken.Create(body.Username)
	if err != nil {
		return nil, fmt.Errorf("access token err: %v", err), nil
	}

	return token, nil, nil
}

func (s *Service) checkUser(body *loginRequest) error {
	user, err := s.Repo.getUser(body.Username)
	if err != nil {
		return err
	}

	return ValidatePassword(body.Password, user.Password)
}

func (s *Service) login(request *http.Request) (token *string, serverErr error, clientErr error) {
	var body loginRequest
	err := json.NewDecoder(request.Body).Decode(&body)
	if err != nil {
		return nil, err, nil
	}

	err = s.checkUser(&body)
	if err != nil {
		log.Printf("[Service] login ERR: %v\n", err)
		return nil, nil, fmt.Errorf("incorrect username or password")
	}

	token, err = accessToken.Create(body.Username)
	if err != nil {
		return nil, err, nil
	}

	return token, nil, nil
}
