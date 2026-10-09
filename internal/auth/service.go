package auth

import (
	"fmt"
	"net/http"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/vilmis04/auth-proxy/internal/accessToken"
	"golang.org/x/crypto/bcrypt"
)

const (
	minUsernameLen = 3
	maxUsernameLen = 32
	minPasswordLen = 10
	// bcrypt ignores everything after the first 72 bytes.
	maxPasswordLen = 72
)

var errInvalidCredentials = &ClientError{Status: http.StatusUnauthorized, Msg: "incorrect username or password"}

// dummyHash is compared against when the user does not exist, so that
// unknown and known users cost the same bcrypt work.
var dummyHash = sync.OnceValue(func() []byte {
	hash, err := bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing"), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return hash
})

type Service struct {
	repo    UserRepo
	signer  *accessToken.Signer
	compare func(password string, hashedPassword []byte) error
}

func NewService(repo UserRepo, signer *accessToken.Signer) *Service {
	return &Service{repo: repo, signer: signer, compare: ValidatePassword}
}

func (s *Service) getIsAuthenticated(token string) (*string, error) {
	user, err := s.signer.Validate(token)
	if err != nil {
		return nil, err
	}
	if user == nil || *user == "" {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

func clientErr(status int, format string, args ...any) *ClientError {
	return &ClientError{Status: status, Msg: fmt.Sprintf(format, args...)}
}

func validateUsername(username string) error {
	n := utf8.RuneCountInString(username)
	if !utf8.ValidString(username) || n < minUsernameLen || n > maxUsernameLen {
		return clientErr(http.StatusBadRequest, "username must be %d-%d characters", minUsernameLen, maxUsernameLen)
	}
	first, _ := utf8.DecodeRuneInString(username)
	last, _ := utf8.DecodeLastRuneInString(username)
	if unicode.IsSpace(first) || unicode.IsSpace(last) {
		return clientErr(http.StatusBadRequest, "username must not start or end with whitespace")
	}
	for _, r := range username {
		if r != ' ' && !unicode.IsPrint(r) {
			return clientErr(http.StatusBadRequest, "username contains invalid characters")
		}
	}

	return nil
}

func validatePassword(password string) error {
	if len(password) < minPasswordLen || len(password) > maxPasswordLen {
		return clientErr(http.StatusBadRequest, "password must be %d-%d bytes", minPasswordLen, maxPasswordLen)
	}

	return nil
}

func (s *Service) signUp(body signUpRequest) (*string, error) {
	if err := validateUsername(body.Username); err != nil {
		return nil, err
	}
	if err := validatePassword(body.Password); err != nil {
		return nil, err
	}
	if body.Password != body.RepeatPassword {
		return nil, clientErr(http.StatusBadRequest, "passwords do not match")
	}

	hashedPassword, err := HashPassword(body.Password)
	if err != nil {
		return nil, fmt.Errorf("hashing: %w", err)
	}
	err = s.repo.CreateUser(body.Username, *hashedPassword)
	if err == ErrUsernameTaken {
		return nil, clientErr(http.StatusConflict, "%v", ErrUsernameTaken)
	}
	if err != nil {
		return nil, fmt.Errorf("user creation: %w", err)
	}

	token, err := s.signer.Create(body.Username)
	if err != nil {
		return nil, fmt.Errorf("access token: %w", err)
	}

	return token, nil
}

func (s *Service) checkUser(body loginRequest) error {
	if utf8.RuneCountInString(body.Username) > maxUsernameLen || len(body.Password) > maxPasswordLen {
		return errInvalidCredentials
	}

	user, err := s.repo.GetUser(body.Username)
	if err == ErrUserNotFound {
		// Same bcrypt cost as a real check, so response time does not reveal which usernames exist.
		_ = s.compare(body.Password, dummyHash())
		return errInvalidCredentials
	}
	if err != nil {
		return fmt.Errorf("user lookup: %w", err)
	}
	if s.compare(body.Password, user.Password) != nil {
		return errInvalidCredentials
	}

	return nil
}

func (s *Service) login(body loginRequest) (*string, error) {
	if err := s.checkUser(body); err != nil {
		return nil, err
	}

	token, err := s.signer.Create(body.Username)
	if err != nil {
		return nil, fmt.Errorf("access token: %w", err)
	}

	return token, nil
}
