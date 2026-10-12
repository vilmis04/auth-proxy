package auth

import "errors"

type User struct {
	Username string `json:"username"`
	Password []byte `json:"password"`
}

type signUpRequest struct {
	Username       string `json:"username"`
	RepeatPassword string `json:"repeatPassword"`
	Password       string `json:"password"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UserResponse is the body of successful sign-up and login responses.
// The token itself is only sent in the HttpOnly cookie.
type UserResponse struct {
	Username string `json:"username"`
}

var (
	ErrUsernameTaken = errors.New("username is already taken")
	ErrUserNotFound  = errors.New("user not found")
)

// ClientError is a failure caused by the request; Status is the HTTP status to return.
type ClientError struct {
	Status int
	Msg    string
}

func (e *ClientError) Error() string {
	return e.Msg
}

// UserRepo is the storage the auth service depends on.
type UserRepo interface {
	// CreateUser returns ErrUsernameTaken when the username already exists.
	CreateUser(username string, passwordHash []byte) error
	// GetUser returns ErrUserNotFound when there is no such user.
	GetUser(username string) (*User, error)
}
