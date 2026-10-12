package accessToken

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testKey = "0123456789abcdef0123456789abcdef"

func newTestSigner(t *testing.T) *Signer {
	t.Helper()
	s, err := NewSigner(testKey, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestNewSignerRejectsWeakKeys(t *testing.T) {
	for name, key := range map[string]string{
		"empty": "",
		"short": strings.Repeat("a", MinKeyLength-1),
	} {
		if _, err := NewSigner(key, time.Hour); err == nil {
			t.Errorf("%s key: expected error", name)
		}
	}
	if _, err := NewSigner(strings.Repeat("a", MinKeyLength), time.Hour); err != nil {
		t.Errorf("32-byte key should be accepted: %v", err)
	}
}

func TestNewSignerRejectsNonPositiveTTL(t *testing.T) {
	for _, ttl := range []time.Duration{0, -time.Minute} {
		if _, err := NewSigner(testKey, ttl); err == nil {
			t.Errorf("ttl %v: expected error", ttl)
		}
	}
}

func TestCreateAndValidate(t *testing.T) {
	s := newTestSigner(t)
	token, err := s.Create("test")
	if err != nil {
		t.Fatal(err)
	}
	user, err := s.Validate(*token)
	if err != nil {
		t.Fatal(err)
	}
	if *user != "test" {
		t.Errorf("expected: test, received: %v", *user)
	}
}

func TestCreateSetsIssuedAtAndExpiry(t *testing.T) {
	s := newTestSigner(t)
	token, _ := s.Create("test")
	claims := jwt.MapClaims{}
	if _, _, err := jwt.NewParser().ParseUnverified(*token, claims); err != nil {
		t.Fatal(err)
	}
	iat, _ := claims.GetIssuedAt()
	exp, _ := claims.GetExpirationTime()
	if iat == nil || exp == nil {
		t.Fatalf("iat and exp must be set, got %v", claims)
	}
	if got := exp.Sub(iat.Time); got != time.Hour {
		t.Errorf("expected lifetime 1h, got %v", got)
	}
}

func TestValidateRejectsExpiredToken(t *testing.T) {
	s := newTestSigner(t)
	s.now = func() time.Time { return time.Now().Add(-2 * time.Hour) }
	token, err := s.Create("test")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := newTestSigner(t).Validate(*token); err == nil {
		t.Error("expected expired token to be rejected")
	}
}

func TestValidateRejectsTokenWithoutExpiry(t *testing.T) {
	signed, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "test"}).SignedString([]byte(testKey))
	if _, err := newTestSigner(t).Validate(signed); err == nil {
		t.Error("expected token without exp to be rejected")
	}
}

func TestValidateRejectsWrongAlgorithm(t *testing.T) {
	claims := jwt.MapClaims{
		"sub": "test",
		"exp": jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
	s := newTestSigner(t)

	hs384, _ := jwt.NewWithClaims(jwt.SigningMethodHS384, claims).SignedString([]byte(testKey))
	if _, err := s.Validate(hs384); err == nil {
		t.Error("expected HS384 token to be rejected")
	}

	none, _ := jwt.NewWithClaims(jwt.SigningMethodNone, claims).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if _, err := s.Validate(none); err == nil {
		t.Error("expected alg=none token to be rejected")
	}
}

func TestValidateRejectsWrongKey(t *testing.T) {
	other, _ := NewSigner(strings.Repeat("x", MinKeyLength), time.Hour)
	token, _ := other.Create("test")
	if _, err := newTestSigner(t).Validate(*token); err == nil {
		t.Error("expected token signed with another key to be rejected")
	}
}

func TestValidateRejectsEmptySubject(t *testing.T) {
	s := newTestSigner(t)
	token, _ := s.Create("")
	if _, err := s.Validate(*token); err == nil {
		t.Error("expected token without subject to be rejected")
	}
}
