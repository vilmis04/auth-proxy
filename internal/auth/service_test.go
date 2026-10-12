package auth

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/vilmis04/auth-proxy/internal/accessToken"
)

const testKey = "0123456789abcdef0123456789abcdef"

// fakeRepo stores usernames exactly as received, like the real table.
type fakeRepo struct {
	users map[string][]byte
	err   error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{users: map[string][]byte{}}
}

func (f *fakeRepo) CreateUser(username string, passwordHash []byte) error {
	if f.err != nil {
		return f.err
	}
	if _, ok := f.users[username]; ok {
		return ErrUsernameTaken
	}
	f.users[username] = passwordHash
	return nil
}

func (f *fakeRepo) GetUser(username string) (*User, error) {
	if f.err != nil {
		return nil, f.err
	}
	hash, ok := f.users[username]
	if !ok {
		return nil, ErrUserNotFound
	}
	return &User{Username: username, Password: hash}, nil
}

func newTestService(t *testing.T, repo UserRepo) *Service {
	t.Helper()
	signer, err := accessToken.NewSigner(testKey, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return NewService(repo, signer)
}

func signUpBody(username, password string) signUpRequest {
	return signUpRequest{Username: username, Password: password, RepeatPassword: password}
}

func statusOf(t *testing.T, err error) int {
	t.Helper()
	var ce *ClientError
	if !errors.As(err, &ce) {
		t.Fatalf("expected ClientError, got %v", err)
	}
	return ce.Status
}

func TestSignUpAndLoginWithSpecialCharacters(t *testing.T) {
	for _, username := range []string{`a&b`, `<b>x</b>`, `o'neil`, `say "hi"`, `a&amp;b`} {
		service := newTestService(t, newFakeRepo())
		if _, err := service.signUp(signUpBody(username, "correct horse")); err != nil {
			t.Fatalf("%q sign-up: %v", username, err)
		}
		token, err := service.login(loginRequest{Username: username, Password: "correct horse"})
		if err != nil {
			t.Fatalf("%q login: %v", username, err)
		}
		user, err := service.getIsAuthenticated(*token)
		if err != nil || *user != username {
			t.Errorf("%q: token subject = %v, err = %v", username, user, err)
		}
	}
}

func TestSignUpStoresUsernameUnescaped(t *testing.T) {
	repo := newFakeRepo()
	if _, err := newTestService(t, repo).signUp(signUpBody("a&b", "correct horse")); err != nil {
		t.Fatal(err)
	}
	if _, ok := repo.users["a&b"]; !ok {
		t.Errorf("expected raw username to be stored, got %v", repo.users)
	}
}

func TestSignUpDuplicateUsername(t *testing.T) {
	service := newTestService(t, newFakeRepo())
	if _, err := service.signUp(signUpBody("a&b", "correct horse")); err != nil {
		t.Fatal(err)
	}
	_, err := service.signUp(signUpBody("a&b", "another password"))
	if got := statusOf(t, err); got != http.StatusConflict {
		t.Errorf("expected 409, got %d", got)
	}
}

func TestSignUpValidation(t *testing.T) {
	service := newTestService(t, newFakeRepo())
	tests := map[string]signUpRequest{
		"short username":    signUpBody("ab", "correct horse"),
		"long username":     signUpBody(strings.Repeat("a", 33), "correct horse"),
		"leading space":     signUpBody(" abc", "correct horse"),
		"trailing space":    signUpBody("abc ", "correct horse"),
		"control character": signUpBody("ab\x00c", "correct horse"),
		"newline":           signUpBody("ab\nc", "correct horse"),
		"short password":    signUpBody("alice", strings.Repeat("p", 9)),
		"long password":     signUpBody("alice", strings.Repeat("p", 73)),
		"password mismatch": {Username: "alice", Password: "correct horse", RepeatPassword: "other horse"},
	}
	for name, body := range tests {
		_, err := service.signUp(body)
		if got := statusOf(t, err); got != http.StatusBadRequest {
			t.Errorf("%s: expected 400, got %d", name, got)
		}
	}

	if _, err := service.signUp(signUpBody(strings.Repeat("ä", 32), strings.Repeat("p", 72))); err != nil {
		t.Errorf("boundary values should be accepted: %v", err)
	}
}

func TestSignUpRepoFailureIsServerError(t *testing.T) {
	repo := newFakeRepo()
	repo.err = errors.New("db down")
	_, err := newTestService(t, repo).signUp(signUpBody("alice", "correct horse"))
	var ce *ClientError
	if err == nil || errors.As(err, &ce) {
		t.Errorf("expected a non-client error, got %v", err)
	}
}

func TestLoginWrongPasswordAndUnknownUserLookAlike(t *testing.T) {
	service := newTestService(t, newFakeRepo())
	if _, err := service.signUp(signUpBody("alice", "correct horse")); err != nil {
		t.Fatal(err)
	}

	_, wrongPassword := service.login(loginRequest{Username: "alice", Password: "wrong password"})
	_, unknownUser := service.login(loginRequest{Username: "nobody", Password: "wrong password"})
	if statusOf(t, wrongPassword) != http.StatusUnauthorized || statusOf(t, unknownUser) != http.StatusUnauthorized {
		t.Fatal("expected 401 for both")
	}
	if wrongPassword.Error() != unknownUser.Error() {
		t.Errorf("messages differ: %q vs %q", wrongPassword, unknownUser)
	}
}

func TestLoginRunsBcryptCompareForUnknownUser(t *testing.T) {
	service := newTestService(t, newFakeRepo())
	calls := 0
	service.compare = func(string, []byte) error {
		calls++
		return errors.New("mismatch")
	}

	if _, err := service.login(loginRequest{Username: "nobody", Password: "whatever password"}); err == nil {
		t.Fatal("expected error")
	}
	if calls != 1 {
		t.Errorf("expected one dummy compare, got %d", calls)
	}
}

func TestLoginRepoFailureIsServerError(t *testing.T) {
	repo := newFakeRepo()
	repo.err = errors.New("db down")
	_, err := newTestService(t, repo).login(loginRequest{Username: "alice", Password: "correct horse"})
	var ce *ClientError
	if err == nil || errors.As(err, &ce) {
		t.Errorf("a database failure must not look like bad credentials, got %v", err)
	}
}

func TestLoginRejectsOversizedInput(t *testing.T) {
	service := newTestService(t, newFakeRepo())
	_, err := service.login(loginRequest{Username: "alice", Password: strings.Repeat("p", 100)})
	if statusOf(t, err) != http.StatusUnauthorized {
		t.Error("expected 401 for an over-long password")
	}
}
