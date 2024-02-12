package auth

import "github.com/vilmis04/auth-proxy/internal/jwtUtils"

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) getIsAuthorized(jwtString string) bool {
	user, err := jwtUtils.ValidateJWT(jwtString)
	if err != nil {
		return false
	}
	if user == nil || *user == "" {
		return false
	}

	return true
}
